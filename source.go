package pgpkg

// This file contains various ways to create source data for a project.
// Mostly you probably want to look in project.go.

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Source interface {
	// FS returns the filesystem associated with the given source.
	FS() (fs.FS, error)

	// Sub returns a source representing a subdirectory within the source.
	Sub(dir string) (Source, error)

	// Location should return the actual path for a source, taking account
	// any subpaths that have been extracted from it. This is going to require a different
	// format and handling for directories, embeds, ZIPs, and other objects.
	// TODO
	Location() string
}

type FSSource struct {
	location string
	fs       fs.FS
}

func NewFSSource(efs fs.FS, dir string) (*FSSource, error) {
	root, err := fs.Sub(efs, dir)
	if err != nil {
		return nil, err
	}
	return &FSSource{fs: root, location: "embedded:" + dir}, nil
}

func (f *FSSource) FS() (fs.FS, error) {
	return f.fs, nil
}

func (f *FSSource) Sub(path string) (Source, error) {
	subfs, err := fs.Sub(f.fs, path)
	if err != nil {
		return nil, err
	}

	return &FSSource{location: filepath.Join(f.location, path), fs: subfs}, nil
}

func (f *FSSource) Location() string {
	return f.location
}

type DirSource struct {
	location string
	fs       fs.FS
}

func NewDirSource(path string) *DirSource {
	return &DirSource{location: path, fs: os.DirFS(path)}
}

func (ps *DirSource) FS() (fs.FS, error) {
	return ps.fs, nil
}

func (ps *DirSource) Sub(path string) (Source, error) {
	subfs, err := fs.Sub(ps.fs, path)
	if err != nil {
		return nil, err
	}

	return &DirSource{location: filepath.Join(ps.location, path), fs: subfs}, nil
}

func (ps *DirSource) Location() string {
	return ps.location
}

type ZipByteSource struct {
	location string
	fs       fs.FS
}

// NewZipByteSource creates a new ZIP source from a byte slice.
func NewZipByteSource(zipBytes []byte) (*ZipByteSource, error) {
	byteReader := bytes.NewReader(zipBytes)
	zipfs, err := zip.NewReader(byteReader, int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("unable to read ZIP data: %w", err)
	}

	return &ZipByteSource{location: "embedded:", fs: zipfs}, nil
}

// NewZipPathSource creates a new ZIP source from a filesystem path.
// This reads the whole ZIP file into memory and returns a ZipByteSource.
func NewZipPathSource(path string) (*ZipByteSource, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewZipByteSource(b)
}

func (zs *ZipByteSource) FS() (fs.FS, error) {
	return zs.fs, nil
}

// Sub returns a subtree of a ZipByteSource, as a ZipByteSource.
func (zs *ZipByteSource) Sub(dir string) (Source, error) {
	newFs, err := fs.Sub(zs.fs, dir)
	if err != nil {
		return nil, err
	}

	return &ZipByteSource{fs: newFs, location: filepath.Join(zs.location, dir)}, nil
}

func (zs *ZipByteSource) Location() string {
	return zs.location
}

// NewSource returns a Source based on the given filesystem path.
// If the path name ends in ".zip", NewSource will return a ZipByteSource.
// Otherwise, NewSource returns a DirSource.
func NewSource(pkgPath string) (Source, error) {
	if strings.HasSuffix(pkgPath, ".zip") {
		return NewZipPathSource(pkgPath)
	} else {
		return NewDirSource(pkgPath), nil
	}
}
