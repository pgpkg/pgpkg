package pgpkg

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/lib/pq"
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

	config *configType
}

// Load the settings
type configType struct {
	Package    string
	Schema     string
	Extensions []string
	Uses       []string
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

func (p *Package) logQuery(query string, args []any) {
	if !p.Options.Verbose {
		return
	}

	if args == nil || len(args) == 0 {
		fmt.Println(query)
	} else {
		fmt.Println(query, args)
	}
}

func (p *Package) Exec(tx *sql.Tx, query string, args ...any) (sql.Result, error) {
	p.logQuery(query, args)
	return tx.Exec(query, args...)
}

func (p *Package) QueryRow(tx *sql.Tx, query string, args ...any) *sql.Row {
	p.logQuery(query, args)
	return tx.QueryRow(query, args...)
}

func (p *Package) setRole(tx *sql.Tx) {
	_, err := p.Exec(tx, fmt.Sprintf("set role \"%s\"", p.SchemaName))
	if err != nil {
		panic(fmt.Errorf("unable to change to role %s: %w", p.SchemaName, err))
	}
}

func (p *Package) resetRole(tx *sql.Tx) {
	_, err := p.Exec(tx, fmt.Sprintf("reset role"))
	if err != nil {
		panic(fmt.Errorf("unable to change to role %s: %w", p.SchemaName, err))
	}
}

func (p *Package) hasRole(tx *sql.Tx) bool {
	var roleCount int
	row := p.QueryRow(tx, "select count(*) from pg_roles where rolname=$1", p.SchemaName)
	err := row.Scan(&roleCount)
	if err != nil {
		panic(err)
	}
	return roleCount == 1
}

func (p *Package) createSchema(tx *sql.Tx) error {
	LogQuieter()
	defer LogLouder()

	if !p.hasRole(tx) {
		_, err := p.Exec(tx, fmt.Sprintf("create role \"%s\"", p.SchemaName))
		if err != nil {
			return fmt.Errorf("unable to create role %s: %w", p.SchemaName, err)
		}
	}

	_, err := p.Exec(tx, fmt.Sprintf("create schema if not exists \"%s\" authorization \"%s\"", p.SchemaName, p.SchemaName))
	if err != nil {
		return fmt.Errorf("unable to create schema %s: %w", p.SchemaName, err)
	}

	exts := p.config.Extensions
	if exts != nil {
		for _, ext := range p.config.Extensions {
			if _, err = p.Exec(tx, fmt.Sprintf("create extension if not exists \"%s\" with schema public", ext)); err != nil {
				return fmt.Errorf("unable to create package extension %s: %w", ext, err)
			}
		}
	}

	return nil
}

// Register this package in the pgpkg.pkg table.
func (p *Package) register(tx *sql.Tx) error {
	_, err := p.Exec(tx, "insert into pgpkg.pkg (pkg, schema_name, uses) values ($1, $2, $3) "+
		"on conflict (pkg) do update set schema_name=excluded.schema_name, uses=excluded.uses",
		p.Name, p.SchemaName, pq.Array(p.config.Uses))

	return err
}

func (p *Package) grantPackage(tx *sql.Tx, pkgName string) error {
	var schemaName string
	r := p.QueryRow(tx, "select schema_name from pgpkg.pkg where pkg=$1", pkgName)
	if err := r.Scan(&schemaName); err != nil {
		return err
	}

	if _, err := p.Exec(tx, fmt.Sprintf(`grant usage on schema "%s" to "%s"`, schemaName, p.SchemaName)); err != nil {
		return err
	}

	if _, err := p.Exec(tx, fmt.Sprintf(`grant execute on all functions in schema "%s" to "%s"`, schemaName, p.SchemaName)); err != nil {
		return err
	}

	if _, err := p.Exec(tx, fmt.Sprintf(`grant select, update, insert, references on all tables in schema "%s" to "%s"`, schemaName, p.SchemaName)); err != nil {
		return err
	}

	if _, err := p.Exec(tx, fmt.Sprintf(`grant usage on all sequences in schema "%s" to "%s"`, schemaName, p.SchemaName)); err != nil {
		return err
	}

	return nil
}

// Allow this package to access the packages in the Uses clause of the definition.
func (p *Package) grant(tx *sql.Tx) error {
	if p.config.Uses == nil {
		return nil
	}

	for _, pkg := range p.config.Uses {
		if err := p.grantPackage(tx, pkg); err != nil {
			return err
		}
	}

	return nil
}

func (p *Package) Apply(tx *sql.Tx) error {

	// Stop any other pgpkg process from running simultaneously.
	if _, err := tx.Exec("select pg_advisory_xact_lock(hashtext('pgpkg'))"); err != nil {
		return fmt.Errorf("pgpkg: unable to obtain package lock: %w", err)
	}

	err := p.createSchema(tx)
	if err != nil {
		return err
	}

	if p.API != nil {
		err = p.API.Parse()
		if err != nil {
			return err
		}

		// This runs as pgpkg user since it's accessing pgpkg tables
		// and deleting stuff from the schema.
		err = p.API.purge(tx)
		if err != nil {
			return err
		}

	} else {
		if p.Options.Verbose {
			fmt.Fprintf(os.Stderr, "note: %s: no API defined\n", p.Name)
		}
	}

	// Grant access to any schema declared in the Uses section of the TOML.
	if err = p.grant(tx); err != nil {
		return err
	}

	if p.Schema != nil {
		// Load the migration state outside the schema role.
		if err = p.Schema.loadMigrationState(tx); err != nil {
			return err
		}

		p.setRole(tx)

		if err = p.Schema.Apply(tx); err != nil {
			return err
		}

		p.resetRole(tx)

		// Save the migrated state, also outside the schema role
		if err = p.Schema.saveMigrationState(tx); err != nil {
			return err
		}
	} else {
		if p.Options.Verbose {
			fmt.Fprintf(os.Stderr, "note: %s: no schema defined\n", p.Name)
		}
	}

	if p.API != nil {
		p.setRole(tx)
		if err = p.API.Apply(tx); err != nil {
			return err
		}
		p.resetRole(tx)

		if err = p.API.updateState(tx); err != nil {
			return err
		}
	}

	if err = p.register(tx); err != nil {
		return err
	}

	if p.Tests != nil {
		p.setRole(tx)
		if err := p.Tests.Run(tx); err != nil {
			return err
		}
		p.resetRole(tx)
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
		config:     &config,
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
