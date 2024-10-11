package pgpkg

// A schema is a kind of bundle that implements sequential migrations. It executes statements
// in a strict, specific order.
//
// Build units are identified by their filename within the package, which
// enables us to determine if they have already been run. When new build units
// (ie, file) are added to a schema, they are executed in order.

import (
	"bufio"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Schema struct {
	*Bundle
	migrationDir   string          // root of migration directory (ie, the location of @migration.pgpkg)
	migrationIndex []string        // list of paths that need to be migrated, in order
	migrationPaths map[string]bool // list of paths that need to be migrated, as a map.
	migrationState map[string]bool // set of paths that have already been migrated (loaded from DB)
	migratedState  map[string]bool // set of paths that have been newly migrated
}

func NewSchema(p *Package) *Schema {
	return &Schema{
		Bundle:         p.newBundle(),
		migrationPaths: make(map[string]bool),
	}
}

func (s *Schema) PrintInfo(w InfoWriter) {
	w.Print("Schema bundle", s.migrationDir)
	s.Bundle.PrintInfo(w)
}

func (s *Schema) ApplyUnit(tx *PkgTx, u *Unit) error {
	// unfortunately parser errors return almost no information, so the best
	// we can do is identify the build unit. This seems to be a problem with
	// pg_query_go rather than the underlying PG parser itself.
	if err := u.Parse(); err != nil {
		return fmt.Errorf("unable to upgrade schema: %w", err)
	}

	for _, stmt := range u.Statements {
		_, err := stmt.Try(tx)
		if err != nil {
			return fmt.Errorf("unable to upgrade schema: %w", err)
		}
	}

	return nil
}

func (s *Schema) loadMigrationState(tx *PkgTx) error {
	migrationState := make(map[string]bool)

	// Grab the list of updates that have already been performed
	// This check is disabled when pgpkg decides it needs to self-install.
	if !s.Package.bootstrapSchema {
		migrations, err := tx.Query("select path from pgpkg.migration where pkg=$1", s.Package.Name)
		if err != nil {
			return fmt.Errorf("unable to get migration status: %w", err)
		}

		for migrations.Next() {
			var path string
			if err = migrations.Scan(&path); err != nil {
				return fmt.Errorf("unexpected error: %w", err)
			}

			migrationName := filepath.Base(path)
			migrationState[migrationName] = true
		}
	}

	s.migrationState = migrationState
	return nil
}

func (s *Schema) saveMigrationState(tx *PkgTx) error {
	// Update the pgpkg.migration table to reflect the migration state.
	for path := range s.migratedState {
		if _, err := tx.Exec("insert into pgpkg.migration (pkg, path) values ($1, $2)", s.Package.Name, path); err != nil {
			return fmt.Errorf("unable to save migration state: %w", err)
		}
	}
	return nil
}

// Apply executes the schema statements in order.
func (s *Schema) Apply(tx *PkgTx) error {
	if s.migrationState == nil {
		panic("please call loadMigrationState before calling Apply")
	}

	var err error

	// keep track of the migrations performed, by name.
	migratedState := make(map[string]bool)

	for _, path := range s.migrationIndex {
		unitPath := filepath.Join(s.migrationDir, path)

		// Migrations are identified only by the filename, which means users can
		// refactor and reorganise their file tree without worrying.
		migrationName := filepath.Base(unitPath)

		if !s.migrationState[migrationName] {
			unit, ok := s.getUnit(unitPath)
			if !ok {
				return fmt.Errorf("error: unit not found: %s", unitPath)
			}

			err = s.ApplyUnit(tx, unit)
			if err != nil {
				return err
			}

			s.Package.StatMigrationCount++
			migratedState[migrationName] = true
		}
	}

	s.migratedState = migratedState

	return nil
}

// Load explicit migrations, based on the config file.
// This checks to make sure that there are not two files with the same name
// on different paths.
func (s *Schema) loadMigrations(migrations []string) error {
	if s.migrationIndex != nil {
		return fmt.Errorf("only one of config.Migration or @migration.pgpkg can be specified")
	}

	uniqueNameMap := make(map[string]bool)

	s.migrationDir = "."
	s.migrationIndex = migrations

	for _, path := range migrations {
		migrationName := filepath.Base(path)
		if uniqueNameMap[migrationName] {
			return fmt.Errorf("duplicate migration name '%s' found in path %s", migrationName, path)
		}
		uniqueNameMap[migrationName] = true

		s.migrationPaths[path] = true
		if err := s.addUnit(path); err != nil {
			return err
		}
	}

	return nil
}

// Load legacy migrations from a directory containing "@migrations.pgpkg"
// You can use either the config file or @migrations.pgpkg, but not both.
// DEPRECATED: please use config.Migrations instead.
func (s *Schema) loadMigrationDir(migrationDir string) error {

	if s.migrationIndex != nil {
		return fmt.Errorf("multiple migrations detected: %s", migrationDir)
	}

	paths, err := s.loadCatalog(migrationDir)
	if err != nil {
		return fmt.Errorf("unable to load migration catalog: %w", err)
	}

	fmt.Println("@migration.pgpkg is deprecated and will be removed soon. Use the Migrations field in config.toml instead")

	s.migrationDir = migrationDir
	s.migrationIndex = paths

	var migrationSet = make(map[string]bool)
	for _, path := range paths {
		s.migrationPaths[path] = true
		migrationSet[path] = true
	}

	// checks that all files in the directory are accounted for.
	// this is a surprisingly common mistake.
	return fs.WalkDir(s.Package.Source, migrationDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(migrationDir, path)
		if err != nil {
			return err
		}

		// Ignore non-SQL files
		if !strings.HasSuffix(relPath, ".sql") {
			return nil
		}

		if !migrationSet[relPath] {
			return fmt.Errorf("warning: %s: not found in %s/%s", relPath, migrationDir, migrationFilename)
		}

		return s.addUnit(path)
	})
}

func (s *Schema) loadCatalog(migrationDir string) ([]string, error) {
	catalog, err := s.Package.Source.Open(filepath.Join(migrationDir, migrationFilename))
	if err != nil {
		return nil, err
	}

	var migrationPaths []string

	scanner := bufio.NewScanner(catalog)
	for scanner.Scan() {
		line := scanner.Text()
		location := strings.TrimSpace(validNames.FindString(line))
		if location != "" && !strings.HasPrefix(location, "#") {
			migrationPaths = append(migrationPaths, location)
		}
	}

	return migrationPaths, nil
}
