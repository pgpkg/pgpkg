package pgpkg

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const SchemaBundleDir = "schema"
const APIBundleDir = "api"
const TestBundleDir = "tests"

var pgIdentifierRegexp = regexp.MustCompile("^[\\pL_][\\pL0-9_$]*$")

// Package represents a single schema in a database. pgpkg
// keeps track of and maintains the objects declared in the
// package, but doesn't touch anything else.
//
// Packages are divided into three bundles, called structure,
// API and tests. Each bundle operates in a unique way.
//
// The database structure is represented by a list of upgrade
// files, which are always executed in order. These files can contain
// any SQL code, but generally contain tables and data type definitions.
//
// The API to the schema is represented by files which contain
// functions, views and triggers. These are managed by pgpkg and
// may be created in any order. pgpkg works out the dependencies between
// them.
//
// Tests are files containing SQL functions, that are executed in order. Tests
// that produce exceptions cause the upgrade to be rolled back.
//
// The structure of a Package is:
//
//    Package -> Bundles (structure, app, tests) -> Units (files) -> Statements
//

type Package struct {
	Name                  string   // canonical, unique name of the pgpkg package
	Location              string   // Location of this package
	Root                  fs.FS    // The filesystem that holds the package
	SchemaName            string   // packages own a single schema
	DisableMigrationCheck bool     // migrate without checking migration table. Allows pgpkg to bootstrap itself.
	Options               *Options // installation options

	StatFuncCount      int // Stat showing the number of functions in the package
	StatViewCount      int // Stat showing the number of views in the package
	StatTriggerCount   int // Stat showing the number of triggers in the package
	StatMigrationCount int // Stat showing how many migration scripts were run
	StatTestCount      int // Stat showing how many tests there are.

	Schema *Schema // probably not a bundle, unless bundles can load on demand
	API    *API
	Tests  *Tests
}

// Bundle represents functional unit of a package, consisting of many Units.
// There are three types of bundles: API, schema and test.
//
// Different bundles have distinct behaviours; structure
// bundles perform upgrades, API bundles replace
// existing code, and test bundles are executed after
// everything else is complete.
type Bundle struct {
	Package *Package         // canonical name of the package.
	Path    string           // Path of this bundle, relative to the Package
	Index   map[string]*Unit // Index of location of each unit.
	Units   []*Unit          // Ordered list of build units within the bundle
}

// Open an arbitrary file from the bundle.
func (b *Bundle) Open(path string) (fs.File, error) {
	return b.Package.Root.Open(filepath.Join(b.Path, path))
}

func (b *Bundle) getUnit(path string) (*Unit, bool) {
	u, ok := b.Index[path]
	return u, ok
}

// addUnit adds a new unit to the package. Note that it doesn't read or parse the unit
// until requested.
func (b *Bundle) addUnit(path string) error {
	unit := &Unit{
		Bundle: b,
		Path:   path,
	}

	b.Units = append(b.Units, unit)
	b.Index[path] = unit
	return nil
}

// Add a bundle. The files within the provided root are the bundle contents.
// Bundles (and the build units they are made up of) are lazily loaded.
func (p *Package) loadBundle(path string) (*Bundle, error) {
	contents, err := p.Root.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %w", path, err)
	}

	dir, ok := contents.(fs.ReadDirFile)
	if !ok {
		return nil, fmt.Errorf("%s is not a directory", path)
	}

	entries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("unable to read directory %s: %w", path, err)
	}

	bundle := &Bundle{
		Path:    path,
		Package: p,
		Index:   make(map[string]*Unit),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return nil, fmt.Errorf("bundle subdirectories are not yet supported: %s/%s", path, entry.Name())
		}

		//unitFile, err := root.Open(entry.Name())
		//if err != nil {
		//	return nil, fmt.Errorf("unable to open %s/%s: %w", path, entry.Name(), err)
		//}

		name := entry.Name()
		if strings.HasSuffix(name, ".sql") {
			err = bundle.addUnit(entry.Name())
			if err != nil {
				return nil, err
			}
		}

	}

	return bundle, nil
}

func (p *Package) Apply(tx *sql.Tx) error {

	// Stop any other pgpkg process from running simultaneously.
	if _, err := tx.Exec("select pg_advisory_xact_lock(hashtext('pgpkg'))"); err != nil {
		return fmt.Errorf("pgpkg: unable to obtain package lock: %w", err)
	}

	LogQuieter()
	_, err := tx.Exec(fmt.Sprintf("create schema if not exists \"%s\"", p.SchemaName))
	LogLouder()
	if err != nil {
		return fmt.Errorf("unable to create schema %s: %w", p.SchemaName, err)
	}

	if p.API != nil {
		err := p.API.Parse()
		if err != nil {
			return err
		}

		err = p.API.purge(tx)
		if err != nil {
			return err
		}
	} else {
		if p.Options.Verbose {
			fmt.Fprintf(os.Stderr, "note: %s: no API defined\n", p.Name)
		}
	}

	if p.Schema != nil {
		err := p.Schema.Apply(tx)
		if err != nil {
			return err
		}
	} else {
		if p.Options.Verbose {
			fmt.Fprintf(os.Stderr, "note: %s: no schema defined\n", p.Name)
		}
	}

	if p.API != nil {
		if err := p.API.Apply(tx); err != nil {
			return err
		}
	}

	if p.Tests != nil {
		if err := p.Tests.Run(tx); err != nil {
			return err
		}
	}

	return nil
}

// LoadPackage reads and parses the entire contents of a package,
// dividing it up into bundles, units and statements.
func LoadPackage(location string, root fs.FS, options *Options) (*Package, error) {

	pkgConfigReader, err := root.Open("pgpkg.toml")
	if err != nil {
		return nil, err
	}

	defer pkgConfigReader.Close()

	// Load the settings
	type configType struct {
		Package string
		Schema  string
	}
	var config configType

	if _, err := toml.NewDecoder(pkgConfigReader).Decode(&config); err != nil {
		return nil, fmt.Errorf("unable to read package config: %w", err)
	}

	if !pgIdentifierRegexp.MatchString(config.Schema) {
		return nil, fmt.Errorf("illegal schema name in pgpkg.toml: %s", config.Schema)
	}

	pkg := &Package{
		Name:       config.Package,
		SchemaName: config.Schema,
		Location:   location,
		Root:       root,
		Options:    options,
	}

	// Load the schema

	pkg.Schema, err = pkg.loadSchema(SchemaBundleDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load schema bundle %s: %w", SchemaBundleDir, err)
	}

	// Load the API definitions

	pkg.API, err = pkg.loadAPI(APIBundleDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load API bundle %s: %w", APIBundleDir, err)
	}

	// Load the tests

	pkg.Tests, err = pkg.loadTests(TestBundleDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load test bundle %s: %w", TestBundleDir, err)
	}

	return pkg, nil
}

func (b *Bundle) Location() string {
	return filepath.Join(b.Package.Location, b.Path)
	//return b.Path
}
