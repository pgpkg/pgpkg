package pgpkg

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

// Cache represents a dependency cache which contains the source of any dependencies listed in
// the "Uses" section of pgpkg.toml.
//
// Users can manually import project dependencies into the project cache with `pgpkg import`.
type Cache interface {
	GetCachedSource(pkgName string) (Source, error)
}

type ReadCache struct {
	fs fs.FS // Filesystem derived from the root directory
}

type WriteCache struct {
	ReadCache
	dir string // root directory of the cache, in local filesystem. Typically rooted at the ".pgpkg" directory.
}

var CachePkgNotFound = errors.New("package not found in cache")

func NewWriteCache(dir string) *WriteCache {
	return &WriteCache{dir: dir, ReadCache: ReadCache{fs: os.DirFS(dir)}}
}

func NewReadCache(cfs fs.FS) *ReadCache {
	return &ReadCache{fs: cfs}
}

func (c *ReadCache) GetCachedSource(pkgName string) (Source, error) {
	pkgFs, err := fs.Sub(c.fs, pkgName)
	if err != nil {
		return nil, fmt.Errorf("error finding subfs: %w", err)
	}

	// Check that a package exists here
	f, err := pkgFs.Open("pgpkg.toml")
	if f != nil {
		// we only got the file descriptor to test for existence of the file;
		// we're not intending to use it here.
		_ = f.Close()
	}

	if err == nil {
		return NewFSSource(pkgFs, ".")
	}

	if os.IsNotExist(err) {
		return nil, CachePkgNotFound
	}

	return nil, err
}

// Import the build units into the cache
func (c *WriteCache) importUnits(bundle *Bundle, cachePath string) error {
	// List of directories we've already created. Note that MkdirAll doesn't
	// return an error if a directory exists, so this cache is here simply to avoid calling that
	// possibly expensive function more than necessary.
	dirs := make(map[string]bool)

	for _, unit := range bundle.Units {
		unitpath := filepath.Join(cachePath, unit.Path)

		if err := unit.Parse(); err != nil {
			return fmt.Errorf("unable to parse %s: %w", unitpath, err)
		}

		unitDir := path.Dir(unitpath)
		_, ok := dirs[unitDir]
		if !ok {
			if err := os.MkdirAll(unitDir, 0777); err != nil {
				return err
			}
			dirs[unitDir] = true
		}

		uw, err := os.Create(unitpath)
		if err != nil {
			return fmt.Errorf("unable to create unit file %s: %w", unitpath, err)
		}

		if _, err := uw.Write([]byte(unit.Source)); err != nil {
			_ = uw.Close()
			return fmt.Errorf("unable to write unit file %s: %w", unitpath, err)
		}

		if err := uw.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Import the migration file.
func (c *WriteCache) importMigration(srcPkg *Package, targetPath string) error {
	srcSchema := srcPkg.Schema
	if len(srcSchema.migrationIndex) == 0 {
		return nil // no migration scripts; nothing to import.
	}

	filename := filepath.Join(targetPath, srcSchema.migrationDir, "/@migration.pgpkg")
	dir := path.Dir(filename)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

	mw, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create migration file %s: %w", filename, err)
	}
	defer mw.Close()

	for _, migtationPath := range srcSchema.migrationIndex {
		if _, err := mw.Write([]byte(migtationPath + "\n")); err != nil {
			return fmt.Errorf("unable to add path to migration file %s: %w", filename, err)
		}
	}

	return nil
}

// RemovePackage removes (deletes) a package from the cache.
func (c *WriteCache) RemovePackage(pkgName string) error {
	targetPath := path.Join(c.dir, pkgName)
	return os.RemoveAll(targetPath)
}

// Import a single project into the cache.
func (c *WriteCache) importPackage(pkg *Package) error {
	targetPath := path.Join(c.dir, pkg.Name)

	// If the package being imported is a dependency (presumably of the imported project), and if it
	// already exists in the cache, refuse to import it - since this could potentially downgrade the
	// existing package.
	if pkg.IsDependency {
		_, err := os.Stat(targetPath)
		if err == nil {
			Stdout.Printf("dependency %s already imported, skipping\n", pkg.Name)
			return nil
		}

		if !os.IsNotExist(err) {
			return err
		}
	} else {
		// If the package being imported is not a dependency then we can assume it's been
		// imported directly, which forces the package to be replaced.
		if err := c.RemovePackage(pkg.Name); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(targetPath, 0777); err != nil {
		return err
	}

	pkgConfigFile, err := os.Create(path.Join(targetPath, "pgpkg.toml"))
	if err != nil {
		return err
	}
	defer pkgConfigFile.Close()

	if err := pkg.config.writeConfig(pkgConfigFile); err != nil {
		return err
	}

	if err := c.importMigration(pkg, targetPath); err != nil {
		return err // FIXME: add context
	}

	if err := c.importUnits(pkg.Schema.Bundle, targetPath); err != nil {
		return err
	}

	if err := c.importUnits(pkg.MOB.Bundle, targetPath); err != nil {
		return err
	}

	if err := c.importUnits(pkg.Tests.Bundle, targetPath); err != nil {
		return err
	}

	return nil
}

// ImportProject imports the given project into the cache. If the project has dependencies,
// these are imported from the child project's cache, unless they are already present in the
// target cache.
func (c *WriteCache) ImportProject(srcProject *Project) error {

	// Resolve dependencies on the target project.
	if err := srcProject.resolveDependencies(); err != nil {
		return err
	}

	// Load the packages before we do anything, in case there are problems.
	if err := srcProject.parseSchemas(); err != nil {
		return err
	}

	for _, pkg := range srcProject.pkgs {
		// don't export pgpkg itself.
		if pkg.Name == "github.com/pgpkg/pgpkg" {
			continue
		}

		if err := c.importPackage(pkg); err != nil {
			return err
		}
	}

	return nil
}
