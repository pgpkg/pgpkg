package pgpkg

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/lib/pq"
	"io/fs"
	"os"
)

// Project represents a group of packages that are to be installed into a single
// database. This struct is responsible for downloading and managing packages,
// and arranging for them to be installed in the correct order.
//
// This type is intended to be the main interface for Go integration with `pgpgk`.
//
// You work with a project by adding the packages you need, and then installing it.
// The `pgpkg` package is always installed automatically.

type Project struct {
	Sources []Source
}

func (p *Project) AddFS(fsList ...fs.FS) {
	for _, fsys := range fsList {
		p.Sources = append(p.Sources, &FSSource{fs: fsys, location: "embedded"})
	}
}

// AddZip adds a ZIP filesystem using the given bytes.
// This is useful when using embedded packages go:embed.
func (p *Project) AddZip(zipByteList ...[]byte) {
	for _, zipBytes := range zipByteList {
		p.Sources = append(p.Sources, &ZipSource{zipBytes: zipBytes, location: "embedded"})
	}
}

// AddPath adds a Path to the project, relative to the working directory.
// The path can refer to a ZIP file or a directory.
func (p *Project) AddPath(paths ...string) {
	for _, path := range paths {
		p.Sources = append(p.Sources, &PathSource{path: path, location: path})
	}
}

// Open opens the given database, installs the packages from the project, and
// returns the database connection. Packages are installed within a single transaction.
// Migrations and tests are applied automatically. Package installation is atomic;
// it either fully succeeds or fails without changing the database.

func (p *Project) Open(options *Options) (*sql.DB, error) {

	dsn := os.Getenv("DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DSN environment variable is not set")
	}

	// Load the packages before we do anything, in case there are problems.
	pkgs, err := p.loadPackages(options)
	if err != nil {
		return nil, err
	}

	base, err := pq.NewConnector(dsn)
	if err != nil {
		return nil, fmt.Errorf("connection to database: %w", err)
	}

	// Wrap the connector to print out notices. Capture the options in the handler.
	connector := pq.ConnectorWithNoticeHandler(base,
		func(err *pq.Error) {
			noticeHandler(options, err)
		})

	db := sql.OpenDB(connector)

	dbtx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", err)
	}

	tx := &PkgTx{
		Tx:      dbtx,
		Verbose: true,
	}

	// Initialise pgpkg itself.
	if err := p.Init(tx, options); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("unable to initialize pgpkg: %w", err)
	}

	for _, pkg := range pkgs {
		if err = pkg.Apply(tx); err != nil {
			_ = tx.Rollback()
			_ = db.Close()
			return nil, fmt.Errorf("unable to install package %s: %w", pkg.Name, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to commit package installation: %w", err)
	}

	return db, nil
}

// Load all the packages from the project, and return them.
func (p *Project) loadPackages(options *Options) ([]*Package, error) {

	var pkgs []*Package

	for _, source := range p.Sources {
		pkgfs, err := source.FS()
		if err != nil {
			return nil, fmt.Errorf("unable to load package %s: %w", source.Location(), err)
		}

		pkg, err := loadPackage(p, source.Location(), pkgfs, options)
		if err != nil {
			return nil, err
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

//go:embed pgpkg
var pgpkgFS embed.FS

// PGKSchemaName is the name of the pgpkg schema itself.
const PGKSchemaName = "pgpkg"

// Init initialises the pgpkg schema itself. It effectively uses pgpkg's
// migration tools to bookstrap itself.
func (p *Project) Init(tx *PkgTx, options *Options) error {
	var isInitialised int
	err := tx.QueryRow("select count(*) from information_schema.schemata where schema_name = 'pgpkg'").Scan(&isInitialised)
	if err != nil {
		return fmt.Errorf("unable to read schema: %w", err)
	}

	pkg, err := loadPackage(p, "embedded pgpkg", pgpkgFS, options)
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
