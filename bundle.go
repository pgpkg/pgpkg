package pgpkg

import (
	"io/fs"
	"path/filepath"
)

// Bundle represents functional unit of a package, consisting of many Units.
// Bundles are, effectively, a collection of files.
// There are three types of bundles: MOB, schema and test. These are implemented
// by embedding Bundle.
//
// Different bundles have distinct behaviours; structure
// bundles perform upgrades, MOB bundles replace
// existing code, and test bundles are executed after
// everything else is complete.

type Bundle struct {
	Package *Package         // canonical name of the package.
	Path    string           // Path of this bundle, relative to the Package
	Index   map[string]*Unit // Index of location of each unit.
	Units   []*Unit          // Ordered list of build units within the bundle
}

// HasUnits indicates if any build units were found for this bundle.
func (b *Bundle) HasUnits() bool {
	return b.Units != nil && len(b.Units) > 0
}

// Open an arbitrary file from the bundle.
func (b *Bundle) Open(path string) (fs.File, error) {
	return b.Package.Source.Open(filepath.Join(b.Path, path))
}

func (b *Bundle) getUnit(path string) (*Unit, bool) {
	u, ok := b.Index[path]
	return u, ok
}

// addUnit adds a new unit to the package. Note that it doesn't read or parse the unit
// until requested.
func (b *Bundle) addUnit(path string) error {
	if Options.Verbose {
		Verbose.Printf("%s: add unit: %s", b.Package.Name, path)
	}

	unit := &Unit{
		Bundle: b,
		Path:   path,
	}

	b.Units = append(b.Units, unit)
	b.Index[path] = unit
	return nil
}

func (b *Bundle) Location() string {
	return filepath.Join(b.Package.Location, b.Path)
}
