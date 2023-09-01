package pgpkg

import (
	"archive/zip"
	"fmt"
	"path"
	"path/filepath"
)

// Creates a ZIP archive of a project(s).
//
// The emitted ZIP file consists of a directory for each named package.
//

// Write the contents of the migration file.
func writeMigration(zw *zip.Writer, pkgPath string, schema *Schema) error {
	if len(schema.migrationIndex) == 0 {
		return nil // no migration scripts; nothing to import.
	}

	filename := filepath.Join(pkgPath, schema.migrationDir, "/@migration.pgpkg")
	mw, err := zw.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create migration file %s: %w", filename, err)
	}

	for _, p := range schema.migrationIndex {
		if _, err := mw.Write([]byte(p + "\n")); err != nil {
			return fmt.Errorf("unable to add path to migration file %s: %w", filename, err)
		}
	}

	return nil
}

func zipWriteConfig(zw *zip.Writer, pkgPath string, p *Package) error {
	filename := filepath.Join(pkgPath, "pgpkg.toml")
	tw, err := zw.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create toml file %s: %w", filename, err)
	}

	return p.config.writeConfig(tw)
}

func zipWriteUnits(zw *zip.Writer, pkgPath string, bundle *Bundle) error {
	for _, unit := range bundle.Units {
		unitpath := filepath.Join(pkgPath, unit.Path)

		if err := unit.Parse(); err != nil {
			return fmt.Errorf("unable to parse %s: %w", unitpath, err)
		}

		uw, err := zw.Create(unitpath)
		if err != nil {
			return fmt.Errorf("unable to create unit file %s: %w", unitpath, err)
		}

		if _, err := uw.Write([]byte(unit.Source)); err != nil {
			return fmt.Errorf("unable to write unit file %s: %w", unitpath, err)
		}
	}

	return nil
}

func writePackage(zw *zip.Writer, pkg *Package) error {

	var pkgPath string
	if pkg.IsDependency {
		pkgPath = path.Join(".pgpkg", pkg.config.Package)
	} else {
		pkgPath = "."
	}

	if err := zipWriteConfig(zw, pkgPath, pkg); err != nil {
		return err
	}

	if err := writeMigration(zw, pkgPath, pkg.Schema); err != nil {
		return err // FIXME: add context
	}

	if err := zipWriteUnits(zw, pkgPath, pkg.Schema.Bundle); err != nil {
		return err
	}

	if err := zipWriteUnits(zw, pkgPath, pkg.MOB.Bundle); err != nil {
		return err
	}

	if err := zipWriteUnits(zw, pkgPath, pkg.Tests.Bundle); err != nil {
		return err
	}

	return nil
}

func WriteProject(z *zip.Writer, p *Project) error {
	// load any dependencies for the project. This should come from the project's cache.
	if err := p.resolveDependencies(); err != nil {
		return err
	}

	// Load the package contents.
	if err := p.parseSchemas(); err != nil {
		return err
	}

	mainPackageFound := false

	for _, pkg := range p.pkgs {
		// don't export pgpkg itself.
		if pkg.Name == "github.com/pgpkg/pgpkg" {
			continue
		}

		if !pkg.IsDependency {
			if mainPackageFound {
				return fmt.Errorf("found multiple non-dependency packages")
			}
			mainPackageFound = true
		}

		if err := writePackage(z, pkg); err != nil {
			return err
		}
	}

	return nil
}
