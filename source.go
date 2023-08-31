package pgpkg

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Source represents the tree of files in a package; it's basically a wrapper
// around fs.FS, but adds context. Source lets us access filesystems in any
// format, which currently includes filesystems (eg, for use with go:embed),
// ZIP files (for packaging), and local directories.
//
// Sources may include a cache, which could be either a read or write
// cache, depending on the type of source.
type Source interface {
	// FS is implemented by every Source.
	fs.FS

	// Sub returns a source representing a subdirectory within the source.
	Sub(dir string) (Source, error)

	// Location should return the actual path for a source, taking account
	// any subpaths that have been extracted from it. This is going to require a different
	// format and handling for directories, embeds, ZIPs, and other objects.
	Location() string

	// Cache returns the cache for this source, if one exists. You should return
	// a WriteCache from this function if your source supports writing. FIXME.
	Cache() (Cache, error)
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

func (f *FSSource) Open(name string) (fs.File, error) {
	return f.fs.Open(name)
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

func (f *FSSource) Cache() (Cache, error) {
	cache, err := fs.Sub(f.fs, ".pgpkg")
	if err != nil {
		return nil, err
	}

	return NewReadCache(cache), nil
}

type DirSource struct {
	dir string
	fs  fs.FS
}

func NewDirSource(path string) *DirSource {
	return &DirSource{dir: path, fs: os.DirFS(path)}
}

func (ds *DirSource) Sub(path string) (Source, error) {
	subfs, err := fs.Sub(ds.fs, path)
	if err != nil {
		return nil, err
	}

	return &DirSource{dir: filepath.Join(ds.dir, path), fs: subfs}, nil
}

func (ds *DirSource) Location() string {
	return ds.dir
}

func (ds *DirSource) Cache() (Cache, error) {
	cacheDir := path.Join(ds.dir, ".pgpkg")
	i, err := os.Stat(cacheDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(cacheDir, 0700)
		if err != nil {
			return nil, fmt.Errorf("unable to create package cache %s: %w", cacheDir, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("unable to open cache: %w", err)
	} else {
		if !i.IsDir() {
			return nil, fmt.Errorf("package cache %s: not a directory", cacheDir)
		}
	}
	return NewWriteCache(cacheDir), nil
}

func (ds *DirSource) Open(name string) (fs.File, error) {
	return ds.fs.Open(name)
}

// Path returns the path that this DirSource refers to, allowing discovery of the
// underlying filesystem location.
func (ds *DirSource) Path() string {
	return ds.dir
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

// Sub returns a subtree of a ZipByteSource, as a ZipByteSource.
func (zs *ZipByteSource) Sub(dir string) (Source, error) {
	newFs, err := fs.Sub(zs.fs, dir)
	if err != nil {
		return nil, err
	}

	return &ZipByteSource{fs: newFs, location: filepath.Join(zs.location, dir)}, nil
}

func (zs *ZipByteSource) Cache() (Cache, error) {
	cache, err := fs.Sub(zs.fs, ".pgpkg")
	if err != nil {
		return nil, err
	}

	return NewReadCache(cache), nil
}

func (zs *ZipByteSource) Location() string {
	return zs.location
}

func (zs *ZipByteSource) Open(name string) (fs.File, error) {
	return zs.fs.Open(name)
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
