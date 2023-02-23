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
	"regexp"
	"strings"
)

type Schema struct {
	*Bundle
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

// readCatalog reads the contents of a catalog and returns the list of Units
// identified in it. Returns an error if a unit can't be found.
func (s *Schema) readCatalog(reader fs.File) ([]*Unit, error) {
	validNames, _ := regexp.Compile("[^#]*")

	var units []*Unit

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		location := strings.TrimSpace(validNames.FindString(line))
		if location != "" {
			unit, ok := s.getUnit(location)
			if !ok {
				return nil, fmt.Errorf("unit not found: %s", location)
			}
			units = append(units, unit)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to read catalog: %w", err)
	}

	return units, nil
}

// Apply executes the schema statements in order.
func (s *Schema) Apply(tx *sql.Tx) error {

	// The catalog lists schema migration files in strict order.
	catalog, err := s.Bundle.Open("@index.pgpkg")
	if err != nil {
		return err
	}

	units, err := s.readCatalog(catalog)
	if err != nil {
		return err
	}

	migrationStatus := make(map[string]bool)

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
			migrationStatus[path] = true
		}
	}

	// Keep track of the migrations that have been applied.
	var applied []string

	for _, u := range units {
		if !migrationStatus[u.Path] {
			err := s.ApplyUnit(tx, u)
			if err != nil {
				return err
			}

			s.Package.StatMigrationCount++
			applied = append(applied, u.Path)
		}
	}

	// Finally, update the pgpkg.migration table. If this is the first time pgpkg
	// has run, the migration table should now exist.
	for _, path := range applied {
		if _, err = tx.Exec("insert into pgpkg.migration (pkg, path) values ($1, $2)", s.Package.Name, path); err != nil {
			return fmt.Errorf("unable to save migration state: %w", err)
		}
	}

	return nil
}
