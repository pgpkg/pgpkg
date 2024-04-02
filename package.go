package pgpkg

import (
	"fmt"
	"github.com/lib/pq"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const migrationFilename = "@migration.pgpkg"

// Package represents a set of schemas in a database. pgpkg
// keeps track of and maintains the objects declared in the
// package, but doesn't touch anything else.
//
// Packages are divided into three bundles, called schema,
// MOB and tests. Each bundle operates in a unique way.
//
// The database structure is represented by a list of upgrade
// files, which are always executed in order. These files can contain
// any SQL code, but generally contain tables and data type definitions.
//
// The MOB to the schema is represented by files which contain
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
	Project     *Project
	Name        string   // canonical, unique name of the pgpkg package
	Location    string   // Location of this package
	Source      Source   // Source of the package (dir, zip, embedded, ...)
	SchemaNames []string // Packages participate in one or more schemas
	RoleName    string   // Associated role name

	StatFuncCount      int // Stat showing the number of functions in the package
	StatViewCount      int // Stat showing the number of views in the package
	StatTriggerCount   int // Stat showing the number of triggers in the package
	StatMigrationCount int // Stat showing how many migration scripts were run
	StatTestCount      int // Stat showing how many tests there are.

	Schema *Schema
	MOB    *MOB
	Tests  *Tests

	IsDependency    bool // This package was loaded from .pgpkg cache
	bootstrapSchema bool // migrate without checking migration table. Allows pgpkg to bootstrap itself.
	config          *configType
}

func (p *Package) newBundle() *Bundle {
	return &Bundle{
		Path:    "",
		Package: p,
		Index:   make(map[string]*Unit),
	}
}

func (p *Package) setRole(tx *PkgTx) {
	_, err := tx.Exec(fmt.Sprintf("set role \"%s\"", Sanitize(rolePattern, p.RoleName)))
	if err != nil {
		panic(fmt.Errorf("unable to change to role %s: %w", p.RoleName, err))
	}
}

func (p *Package) resetRole(tx *PkgTx) {
	_, err := tx.Exec(fmt.Sprintf("reset role"))
	if err != nil {
		panic(fmt.Errorf("unable to reset to role %s: %w", p.RoleName, err))
	}
}

func (p *Package) hasRole(tx *PkgTx) bool {
	var roleCount int
	row := tx.QueryRow("select count(*) from pg_roles where rolname=$1", p.RoleName)
	err := row.Scan(&roleCount)
	if err != nil {
		panic(err)
	}
	return roleCount == 1
}

func (p *Package) createSchema(tx *PkgTx) error {
	LogQuieter()
	defer LogLouder()

	if !p.hasRole(tx) {
		_, err := tx.Exec(fmt.Sprintf("create role \"%s\"", Sanitize(rolePattern, p.RoleName)))
		if err != nil {
			return fmt.Errorf("unable to create role %s: %w", p.RoleName, err)
		}

		// The user running these scripts may not be a superuser (but must have create role),
		// so we need to extend access to the new role.
		_, err = tx.Exec(fmt.Sprintf("grant \"%s\" to current_user", Sanitize(rolePattern, p.RoleName)))
		if err != nil {
			return fmt.Errorf("unable to grant role %s to current_user: %w", p.RoleName, err)
		}
	}

	for _, schemaName := range p.SchemaNames {
		_, err := tx.Exec(fmt.Sprintf("create schema if not exists \"%s\" authorization \"%s\"",
			Sanitize(schemaPattern, schemaName), Sanitize(rolePattern, p.RoleName)))

		if err != nil {
			return fmt.Errorf("unable to create schema %s: %w", schemaName, err)
		}
	}

	exts := p.config.Extensions
	if exts != nil {
		for _, ext := range p.config.Extensions {
			if _, err := tx.Exec(fmt.Sprintf("create extension if not exists \"%s\" with schema public", Sanitize(extensionPattern, ext))); err != nil {
				return fmt.Errorf("unable to create package extension %s: %w", ext, err)
			}
		}
	}

	return nil
}

// Register this package in the pgpkg.pkg table.
func (p *Package) register(tx *PkgTx) error {
	_, err := tx.Exec("insert into pgpkg.pkg (pkg, schema_names, uses) values ($1, $2, $3) "+
		"on conflict (pkg) do update set schema_names=excluded.schema_names, uses=excluded.uses",
		p.Name, pq.Array(p.SchemaNames), pq.Array(p.config.Uses))

	return err
}

func (p *Package) grantPackage(tx *PkgTx, pkgName string) error {
	var schemaNames []string
	r := tx.QueryRow("select schema_names from pgpkg.pkg where pkg=$1", pkgName)
	if err := r.Scan(pq.Array(&schemaNames)); err != nil {
		return fmt.Errorf("unable to grant access to package %s: %w", pkgName, err)
	}

	for _, schemaName := range schemaNames {
		if _, err := tx.Exec(fmt.Sprintf(`grant usage on schema "%s" to "%s"`,
			Sanitize(schemaPattern, schemaName), Sanitize(rolePattern, p.RoleName))); err != nil {
			return err
		}

		if _, err := tx.Exec(fmt.Sprintf(`grant execute on all functions in schema "%s" to "%s"`,
			Sanitize(schemaPattern, schemaName), Sanitize(rolePattern, p.RoleName))); err != nil {
			return err
		}

		if _, err := tx.Exec(fmt.Sprintf(`grant select, update, insert, references on all tables in schema "%s" to "%s"`,
			Sanitize(schemaPattern, schemaName), Sanitize(rolePattern, p.RoleName))); err != nil {
			return err
		}

		if _, err := tx.Exec(fmt.Sprintf(`grant usage on all sequences in schema "%s" to "%s"`,
			Sanitize(schemaPattern, schemaName), Sanitize(rolePattern, p.RoleName))); err != nil {
			return err
		}
	}

	return nil
}

// grant access to certain parts of the pgpkg package.
func (p *Package) grantPgpkg(tx *PkgTx) error {
	if _, err := tx.Exec(fmt.Sprintf(`grant usage on schema "pgpkg" to "%s"`,
		Sanitize(rolePattern, p.RoleName))); err != nil {
		return err
	}

	if _, err := tx.Exec(fmt.Sprintf(`grant execute on all functions in schema "pgpkg" to "%s"`,
		Sanitize(rolePattern, p.RoleName))); err != nil {
		return err
	}

	return nil
}

// Allow this package to access the packages in the Uses clause of the definition.
func (p *Package) grantUses(tx *PkgTx) error {
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

func (p *Package) Apply(tx *PkgTx) error {

	// Stop any other pgpkg process from running simultaneously.
	if _, err := tx.Exec("select pg_advisory_xact_lock(hashtext('pgpkg'))"); err != nil {
		return fmt.Errorf("pgpkg: unable to obtain package lock: %w", err)
	}

	err := p.createSchema(tx)
	if err != nil {
		return err
	}

	if p.MOB != nil && p.MOB.HasUnits() {
		err = p.MOB.Parse()
		if err != nil {
			return err
		}

		// This runs as pgpkg user since it's accessing pgpkg tables
		// and deleting stuff from the schema.
		err = p.MOB.purge(tx)
		if err != nil {
			return err
		}

	} else {
		if Options.Verbose {
			fmt.Fprintf(os.Stderr, "note: %s: no MOBs defined\n", p.Name)
		}
	}

	// Grant access to functions in pgpkg, e.g. the assertions
	if err = p.grantPgpkg(tx); err != nil {
		return err
	}

	// Grant access to any schema declared in the Uses section of the TOML.
	if err = p.grantUses(tx); err != nil {
		return err
	}

	if p.Schema != nil && p.Schema.HasUnits() {
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
		if Options.Verbose {
			fmt.Fprintf(os.Stderr, "note: %s: no schema defined\n", p.Name)
		}
	}

	if p.MOB != nil && p.MOB.HasUnits() {
		p.setRole(tx)
		if err = p.MOB.Apply(tx); err != nil {
			return err
		}
		p.resetRole(tx)

		if err = p.MOB.updateState(tx); err != nil {
			return err
		}
	}

	if err = p.register(tx); err != nil {
		return err
	}

	if p.Tests != nil && p.Tests.HasUnits() && !Options.SkipTests {
		p.setRole(tx)
		if err := p.Tests.Run(tx); err != nil {
			return err
		}
		p.resetRole(tx)
	}

	if Options.Verbose || Options.Summary {
		Verbose.Printf("%s: installed %d function(s), %d view(s) and %d trigger(s). %d migration(s) needed. %d test(s) run\n",
			p.Name, p.StatFuncCount, p.StatViewCount, p.StatTriggerCount, p.StatMigrationCount, p.StatTestCount)
	}

	return nil
}

// Read the configuration TOML file and update the package accordingly.
// If the package is already configured, it's an error.
func (p *Package) parseConfig(tomlPath string) error {

	if p.config != nil {
		return fmt.Errorf("duplicate configuration found: %s", tomlPath)
	}

	pkgConfigReader, err := p.Source.Open(tomlPath)
	if err != nil {
		return err
	}

	defer pkgConfigReader.Close()

	config, err := parseConfig(pkgConfigReader)
	if err != nil {
		return err
	}

	p.Name = config.Package
	p.SchemaNames = SanitizeSlice(schemaPattern, config.Schemas)
	p.RoleName = Sanitize(rolePattern, "$"+p.Name)
	p.config = config

	return nil
}

var validNames = regexp.MustCompile("[^#]*")

func (p *Package) addUnit(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	name := d.Name()

	// Ignore the pgpkg.toml file.
	if name == "pgpkg.toml" {
		return nil
	}

	// Ignore dot-files other than "." itself
	if name != "." && name[0] == '.' {
		if d.IsDir() {
			return fs.SkipDir
		} else {
			return nil
		}
	}

	if d.IsDir() {
		// If this is a directory, and it contains migrations, then
		// process it with a separate walk().
		if _, err = fs.Stat(p.Source, filepath.Join(path, migrationFilename)); err == nil {
			if err = p.Schema.loadMigrations(path); err != nil {
				return err
			}
			return fs.SkipDir
		}
	}

	if strings.HasSuffix(name, "_test.sql") {
		return p.Tests.addUnit(path)
	}

	if strings.HasSuffix(name, ".sql") {
		return p.MOB.addUnit(path)
	}

	// Files that aren't recognised are just ignored. This lets us mix pgpkg sql with
	// other files.
	return nil
}

// Load the project details - the TOML file - without reading the rest of the schema data.
func readPackage(project *Project, source Source, dir string) (*Package, error) {
	var err error

	if dir != "" {
		if source, err = source.Sub(dir); err != nil {
			return nil, fmt.Errorf("unable to read package: source %s: %w", source, err)
		}
	}

	pkg := &Package{
		Project:  project,
		Source:   source,
		Location: source.Location(),
	}

	if err := pkg.parseConfig("pgpkg.toml"); err != nil {
		return nil, err
	}

	return pkg, nil
}

func (p *Package) readSchema() error {
	p.Schema = &Schema{Bundle: p.newBundle()}
	p.MOB = &MOB{Bundle: p.newBundle()}
	p.Tests = &Tests{Bundle: p.newBundle()}

	// Only walk the directory in which the toml file was found, rather than
	// the entire filesystem provided in pkgFS.
	if err := fs.WalkDir(p.Source, ".", p.addUnit); err != nil {
		return fmt.Errorf("unable to read schema for package %s: %w", p.Name, err)
	}

	return nil
}

func (p *Package) isValidSchema(search string) bool {
	for _, schema := range p.SchemaNames {
		if schema == search {
			return true
		}
	}

	return false
}

// AddUses adds the given package name to the Uses clause of the package.
// Returns false if the package already exists in the Uses clause.
// Note that this does not update the config file; to do this, see WriteConfig.
func (p *Package) AddUses(pkg string) bool {
	uses := p.config.Uses

	// check that it doesn't already exist
	if uses != nil {
		for _, u := range uses {
			if u == pkg {
				return false
			}
		}
	}

	p.config.Uses = append(p.config.Uses, pkg)
	return true
}

func (p *Package) WriteConfig() error {
	// We can only write to this package if it came from a directory.
	dirFS, ok := p.Source.(*DirSource)
	if !ok {
		return fmt.Errorf("package was not loaded from filesystem")
	}

	tempFile := path.Join(dirFS.Path(), "pgpkg-new.toml")
	pkgConfigWriter, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("unable to create config file %s: %w", tempFile, err)
	}

	if err := p.config.writeConfig(pkgConfigWriter); err != nil {
		return fmt.Errorf("unable to write config file %s: %w", tempFile, err)
	}

	if err := pkgConfigWriter.Close(); err != nil {
		return fmt.Errorf("unable to complete config file write to %s: %w", tempFile, err)
	}

	tomlFile := path.Join(dirFS.Path(), "pgpkg.toml")
	if err := os.Rename(tempFile, tomlFile); err != nil {
		return fmt.Errorf("unable to replace existing pgpkg.toml: %w", err)
	}

	return nil
}

func (p *Package) PrintInfo(w InfoWriter) {
	w.Print("Package name", p.Name)
	w.Print("Location", p.Location)
	w.Print("Source", p.Source)
	w.Print("SchemaNames", p.SchemaNames)
	w.Print("RoleName", p.RoleName)

	if p.Schema != nil {
		p.Schema.PrintInfo(w)
	} else {
		w.Println("no schema")
	}

	if p.MOB != nil {
		p.MOB.PrintInfo(w)
	} else {
		w.Println("no managed objects")
	}

	if p.Tests != nil {
		p.Tests.PrintInfo(w)
	} else {
		w.Println("no tests")
	}

}
