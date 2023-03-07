package pgpkg

// This file contains various ways to create source data for a project.
// Mostly you probably want to look in project.go.

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

type Source interface {
	// FS returns the filesystem associated with the given source.
	FS() (fs.FS, error)
	Location() string
}

type FSSource struct {
	location string
	fs       fs.FS
}

func (f *FSSource) FS() (fs.FS, error) {
	return f.fs, nil
}

func (f *FSSource) Location() string {
	return f.location
}

type PathSource struct {
	location string
	path     string
}

func (ps *PathSource) FS() (fs.FS, error) {

	if !strings.HasSuffix(ps.path, ".zip") {
		return os.DirFS(ps.path), nil
	}

	zipFS, err := zip.OpenReader(ps.path)
	if err != nil {
		return nil, fmt.Errorf("unable to open zip archive: %w", err)
	}

	return zipFS, nil
}

func (ps *PathSource) Location() string {
	return ps.location
}

type ZipSource struct {
	location string
	zipBytes []byte
}

func (zs *ZipSource) FS() (fs.FS, error) {
	byteReader := bytes.NewReader(zs.zipBytes)
	zipfs, err := zip.NewReader(byteReader, int64(len(zs.zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("unable to read ZIP data: %w", err)
	}

	return zipfs, nil
}

func (zs *ZipSource) Location() string {
	return zs.location
}
