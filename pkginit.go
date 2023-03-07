package pgpkg

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
)

//go:embed pgpkg
var pgpkgFS embed.FS

// PGKSchemaName is the name of the pgpkg schema itself.
const PGKSchemaName = "pgpkg"

// Init initialises the pgpkg schema itself. It effectively uses pgpkg's
// migration tools to bookstrap itself.
func Init(tx *sql.Tx, options *Options) error {
	var isInitialised int
	err := tx.QueryRow("select count(*) from information_schema.schemata where schema_name = 'pgpkg'").Scan(&isInitialised)
	if err != nil {
		return fmt.Errorf("unable to read schema: %w", err)
	}

	subFS, err := fs.Sub(pgpkgFS, "pgpkg")
	if err != nil {
		return fmt.Errorf("unable to find pgpkg package: %w", err)
	}

	pkg, err := loadPackage("embedded pgpkg", subFS, options)
	if err != nil {
		return fmt.Errorf("unable to load pgpkg package: %w", err)
	}

	if pkg.SchemaName != PGKSchemaName {
		return fmt.Errorf("expected root schema name %s, got %s", PGKSchemaName, pkg.SchemaName)
	}

	// We can force the package to run all the migration scripts without
	// checking if they have been already run. This prevents the migration
	// from trying to lookup database tables before they are created.
	if isInitialised == 0 {
		pkg.bootstrapSchema = true
	}

	// Apply the pgpkg schema itself.
	if err = pkg.Apply(tx); err != nil {
		return fmt.Errorf("unable to create/update pgpkg package: %w", err)
	}

	return nil
}
