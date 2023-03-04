package pgpkg

// Various utility functions for installing packages. These utilities are intended to be called
// directly by the Go command-line.

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"io/fs"
	"os"
	"strings"
)

// FSList generates a slice of filesystems from a list of pathnames.
// Pathnames that end in ".zip" are treated as ZIP files.
func FSList(pkgPaths ...string) ([]fs.FS, error) {

	var fsList []fs.FS

	for _, pkgPath := range pkgPaths {
		if strings.HasSuffix(pkgPath, ".zip") {
			zipFS, err := zip.OpenReader(pkgPath)
			if err != nil {
				return nil, fmt.Errorf("unable to open zip archive: %w", err)
			}

			fsList = append(fsList, zipFS)
		} else {
			fsList = append(fsList, os.DirFS(pkgPath))
		}

	}

	return fsList, nil
}

// Open the database, initialise pgpkg, and install the given packages (if any).
// Returns the database handle. Upgrades are atomic, but will be completed
// (and committed) when this function returns. This also attaches a message
// handler to the database, so we can log messages from RAISE NOTICE to the
// console.
func Open(conninfo string, options *Options, pkgs ...fs.FS) (*sql.DB, error) {
	base, err := pq.NewConnector(conninfo)
	if err != nil {
		return nil, fmt.Errorf("connection to database: %w", err)
	}

	// Wrap the connector to print out notices. Capture the options in the handler.
	connector := pq.ConnectorWithNoticeHandler(base,
		func(err *pq.Error) {
			noticeHandler(options, err)
		})

	db := sql.OpenDB(connector)

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", err)
	}

	// Initialise pgpkg itself.
	if err := Init(tx, options); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("unable to initialize pgpkg: %w", err)
	}

	if pkgs != nil {
		if err = installFS(tx, options, pkgs...); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("unable to commit package installation: %w", err)
	}

	return db, nil
}

// ZIPFS converts a byte array into a ZIP file, so we can
// use it with installFS. This will panic if the conversion
// fails, since it's intended to deal with an embedded filesystem.
func ZIPFS(zipdata []byte) fs.FS {
	byteReader := bytes.NewReader(zipdata)
	zipfs, err := zip.NewReader(byteReader, int64(len(zipdata)))
	if err != nil {
		panic(fmt.Errorf("unable to read ZIP data: %w", err))
	}

	return zipfs
}

// installFS installs a list of pgpkg packages from filesystems. This is intended for
// use with embedded packages, where we can install packages from a variety of filesystems.
//
// Each filesystem must have a "pgpkg" directory in the root. (This restriction will
// be lifted in future versions of pgpkg).
//
// Use ZIPFS() to wrap a []byte array of a ZIP package for embedding external packages.
func installFS(tx *sql.Tx, options *Options, pkgs ...fs.FS) error {
	for index, pkgFS := range pkgs {
		// Load the package
		pkg, err := loadPackage(fmt.Sprintf("InstallFS[%d]", index), pkgFS, options)
		if err != nil {
			return err
		}

		// Apply the package.
		if err = pkg.Apply(tx); err != nil {
			return err
		}
	}

	return nil
}

// Exit prints the error message (with context, if available), and then exits immediately.
func Exit(err error) {
	PrintError(err)
	os.Exit(1)
}

func PrintError(err error) {
	var pkgErr *PKGError
	ok := errors.As(err, &pkgErr)
	if !ok {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}

	pkgErr.PrintRootContext(2)
}
