package pgpkg

// A schema is a kind of bundle that implements sequential migrations. It executes statements
// in a strict, specific order. (FIXME: how this order is defined is TBD)
//
// Build units are identified by their filename within the package, which
// enables us to determine if they have already been run. When new build units
// (ie, file) are added to a schema, they are executed in order.

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Schema struct {
	*Bundle
	migrationIndex []string        // list of paths that need to be migrated, in order
	migrationState map[string]bool // set of paths that have already been migrated (loaded from DB)
	migratedState  map[string]bool // set of paths that have been newly migrated
}

func (p *Package) loadSchema(path string) (*Schema, error) {
	bundle, err := p.loadBundle(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	schema := &Schema{
		Bundle: bundle,
	}

	return schema, nil
}

func (s *Schema) ApplyUnit(tx *sql.Tx, u *Unit) error {
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

func (s *Schema) loadMigrationState(tx *sql.Tx) error {
	migrationState := make(map[string]bool)

	// Grab the list of updates that have already been performed
	// This check is disabled when pgpkg decides it needs to self-install.
	if !s.Package.DisableMigrationCheck {
		migrations, err := tx.Query("select path from pgpkg.migration where pkg=$1", s.Package.Name)
		if err != nil {
			return fmt.Errorf("unable to get migration status: %w", err)
		}

		for migrations.Next() {
			var path string
			if err = migrations.Scan(&path); err != nil {
				return fmt.Errorf("unexpected error: %w", err)
			}
			migrationState[path] = true
		}
	}

	s.migrationState = migrationState
	return nil
}

func (s *Schema) saveMigrationState(tx *sql.Tx) error {
	// Update the pgpkg.migration table to reflect the migration state.
	for path, _ := range s.migratedState {
		if _, err := tx.Exec("insert into pgpkg.migration (pkg, path) values ($1, $2)", s.Package.Name, path); err != nil {
			return fmt.Errorf("unable to save migration state: %w", err)
		}
	}
	return nil
}

// Apply executes the schema statements in order.
func (s *Schema) Apply(tx *sql.Tx) error {
	if s.migrationState == nil {
		panic("please call loadMigrationState before calling Apply")
	}

	var units []*Unit
	var err error

	for _, path := range s.migrationIndex {
		unit, ok := s.getUnit(path)
		if !ok {
			return fmt.Errorf("error: unit not found: %s", path)
		}
		units = append(units, unit)
	}

	// Keep track of the migrations that have been applied.
	migratedState := make(map[string]bool)

	for _, u := range units {
		if !s.migrationState[u.Path] {
			err = s.ApplyUnit(tx, u)
			if err != nil {
				return err
			}

			s.Package.StatMigrationCount++
			migratedState[u.Path] = true
		}
	}

	s.migratedState = migratedState

	return nil
}

func (s *Schema) loadMigrations(migrationDir string) error {

	if s.migrationIndex != nil {
		return fmt.Errorf("multiple migrations detected: %s", migrationDir)
	}

	paths, err := s.loadCatalog(migrationDir)
	if err != nil {
		return fmt.Errorf("unable to load migration catalog: %w", err)
	}

	s.migrationIndex = paths

	var migrationSet = make(map[string]bool)
	for _, path := range paths {
		migrationSet[path] = true
	}

	return fs.WalkDir(s.Package.Root, migrationDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ignore non-SQL files
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}

		if !migrationSet[path] {
			_, _ = fmt.Fprintf(os.Stderr, "warning: %s: not found in %s/@index.pgpkg", path, migrationDir)
			return nil
		}

		return s.addUnit(path)
	})
}

func (s *Schema) loadCatalog(migrationDir string) ([]string, error) {
	catalog, err := s.Package.Root.Open(filepath.Join(migrationDir, "@index.pgpkg"))
	if err != nil {
		return nil, err
	}

	var units []string

	scanner := bufio.NewScanner(catalog)
	for scanner.Scan() {
		line := scanner.Text()
		location := strings.TrimSpace(validNames.FindString(line))
		if location != "" && !strings.HasPrefix(location, "#") {
			units = append(units, filepath.Join(migrationDir, location))
		}
	}

	return units, nil
}
