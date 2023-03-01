package pgpkg

// Various utility functions for installing packages. These utilities are intended to be called
// directly by the Go command-line.

import (
	"archive/zip"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Install installs the packages identified by the given path using the transaction.
// Paths can be directories and/or ZIP files.
// If Install is successful then the database will have been upgraded to include
// the pgpkgs provided.
//
// WARNING: for changes to be persistent, the caller must commit the transaction.
func Install(tx *sql.Tx, options *Options, pkgPaths ...string) error {
	// Initialise pgpkg itself.
	if err := Init(tx, options); err != nil {
		return err
	}

	// Load all the packages.
	for _, pkgPath := range pkgPaths {

		var pkgFS fs.FS
		var err error

		// For the time being, ZIP files need to contain a "pgpkg" directory;
		// it is this directory which is used to build the package representation.
		// Note that in the future, this constraint will be removed.
		if strings.HasSuffix(pkgPath, ".zip") {
			zipFS, err := zip.OpenReader(pkgPath)
			if err != nil {
				return fmt.Errorf("unable to open zip archive: %w", err)
			}

			pkgFS, err = fs.Sub(zipFS, "pgpkg")
			if err != nil {
				return fmt.Errorf("unable to find pgpkg folder in zip archive: %w", err)
			}
		} else {
			pkgFS = os.DirFS(pkgPath)
		}

		if pkgFS == nil {
			return fmt.Errorf("unable to open package %s", pkgPath)
		}

		// Load the package. loadPackage currently expects files in ./api, ./schema and ./tests.
		pkg, err := loadPackage(pkgPath, pkgFS, options)
		if err != nil {
			return err
		}

		// Apply the requested package.
		if err = pkg.Apply(tx); err != nil {
			return err
		}

		if options.Verbose || options.Summary {
			fmt.Printf("%s: installed %d function(s), %d view(s) and %d trigger(s). %d migration(s) needed. %d test(s) run\n",
				pkg.Name, pkg.StatFuncCount, pkg.StatViewCount, pkg.StatTriggerCount, pkg.StatMigrationCount, pkg.StatTestCount)
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
