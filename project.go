package pgpkg

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/lib/pq"
	"io/fs"
)

// Project represents a collection of individual packages that are to be installed into a single
// database. This struct is responsible for tracking the package sources that make up a project,
// including dependencies and caches, and arranging for them to be installed in the correct order.
//
// You work with a project by adding the "sources" you need - which might be directories, ZIP files,
// embedded filesystems, or embedded ZIP binaries. If a project- or search-cache is defined, then
// this will be used to find dependencies.
//
// Once you've added all the sources for your project, p.Open() or p.Migrate() performs the migration.
//
// The `pgpkg` package is always installed automatically, and is never exported.
type Project struct {
	Root    *Package // the root (main package) of this project
	Sources []Source // All package sources, in no particular order.
	pkgs    map[string]*Package
	Cache   *WriteCache // primary cache for this project
	Search  []Cache     // other caches to search for dependencies.
}

func (p *Project) AddEmbeddedFS(f fs.FS, path string) (*Package, error) {
	src, err := NewFSSource(f, path)
	if err != nil {
		return nil, fmt.Errorf("unable to add embedded FS: %s: %w", path, err)
	}
	return p.AddPackage(src, false)
}

// AddPackage adds an individual package to the project.
func (p *Project) AddPackage(source Source, isDependency bool) (*Package, error) {

	p.Sources = append(p.Sources, source)

	pkg, err := readPackage(p, source, "")
	if err != nil {
		return nil, err
	}

	pkg.IsDependency = isDependency

	existing, ok := p.pkgs[pkg.Name]
	if ok {
		return nil, fmt.Errorf("duplicate package %s; found in %s, but also in %s", pkg.Name, existing.Location, pkg.Location)
	}

	p.pkgs[pkg.Name] = pkg
	return pkg, nil
}

func (p *Project) AddSource(src Source) (*Package, error) {
	return p.AddPackage(src, false)
}

// Get a list of the keys for a map, as a string. Used for debugging.
func mapKeys[T any](m map[string]T) string {
	keys := ""
	for k := range m {
		if keys != "" {
			keys = keys + ","
		}
		keys = keys + k
	}
	return keys
}

func (p *Project) installPackages(tx *PkgTx) error {
	// Sort packages by dependencies.
	pkgs, err := p.sortPackages()
	if err != nil {
		return err
	}

	// Install packages in dependency order.
	for _, pkgName := range pkgs {
		pkg := p.pkgs[pkgName]
		if err := pkg.Apply(tx); err != nil {
			return fmt.Errorf("unable to install package %s: %w", pkg.Name, err)
		}
	}

	return nil
}

func (p *Project) addDependency(uses string) error {
	// Has the dependency already been added to the package?
	_, ok := p.pkgs[uses]
	if ok {
		return nil
	}

	caches := []Cache{}

	// search caches, if any, take precedence over project cache.
	if p.Search != nil {
		caches = append(caches, p.Search...)
	}

	if p.Cache != nil {
		caches = append(caches, p.Cache)
	}

	for _, cache := range caches {
		src, err := cache.GetCachedSource(uses)
		if err == CachePkgNotFound {
			continue
		} else if err != nil {
			return err
		}

		if _, err := p.AddPackage(src, true); err != nil {
			return err
		}

		// Package found, nothing more to do.
		return nil
	}

	return fmt.Errorf("dependency not found")
}

// resolveDependencies adds any dependent packages ("Uses") to the project, by looking in
// the cache(s). The way this works is that we look at the Uses clause of each known package,
// and add the named packages to the project's package list by searching the project caches.
// We then look at the newly added packages to find their dependencies, and so on.
func (p *Project) resolveDependencies() error {
	visited := make(map[string]bool) // list of packages that we have already visited.

	// The project package list is a map, so we can't just iterate through it.
	// Instead, we keep going until there's an error or until all the packages
	// have been visited.
	for len(visited) < len(p.pkgs) {
		// the list of packages gets added to as we resolve dependencies, but we only have to check
		// packages we haven't visited before.
		for _, pkg := range p.pkgs {
			if visited[pkg.Name] {
				continue
			}

			visited[pkg.Name] = true
			for _, uses := range pkg.config.Uses {
				if err := p.addDependency(uses); err != nil {
					return fmt.Errorf("unable to add dependency %s of package %s: %w", uses, pkg.Name, err)
				}
			}
		}
	}

	return nil
}

// Open opens the given database, installs the packages from the project, and
// returns the database connection.
//
// Open is the main entry point for pgpkg.
//
// Packages are installed within a single transaction.
// Migrations and tests are applied automatically. Package installation is atomic;
// it either fully succeeds or fails without changing the database.
//
// If this method returns an error, you should call pgpkg.Exit(err) to exit.
// This call checks that the error was significant and will adjust the OS exit
// status accordingly. See pgpkg.Exit() for more details.
//
// If dsn is an empty string, pgpkg will attempt to use the PGPKG_DSN environment
// variable. If PGPKG_DSN is not set, pgpkg will use the usual libpq PG environment variables.
func (p *Project) Open(dsn string) (*sql.DB, error) {
	if err := p.Parse(); err != nil {
		return nil, err
	}

	base, err := pq.NewConnector(dsn)
	if err != nil {
		return nil, fmt.Errorf("connection to database: %w", err)
	}

	// Wrap the connector to print out notices. Capture the options in the handler.
	connector := pq.ConnectorWithNoticeHandler(base,
		func(err *pq.Error) {
			noticeHandler(err)
		})

	db := sql.OpenDB(connector)

	dbtx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", err)
	}

	tx := &PkgTx{
		Tx: dbtx,
	}

	// Initialise pgpkg itself.
	if err := p.Init(tx); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("unable to initialize pgpkg: %w", err)
	}

	if err := p.installPackages(tx); err != nil {
		_ = tx.Rollback()
		_ = db.Close()
		return nil, fmt.Errorf("unable to complete package installation: %w", err)
	}

	if Options.DryRun {
		err = tx.Rollback()
		_ = db.Close()
		if err != nil {
			return nil, err
		}
		return nil, ErrDryRun
	} else {
		err = tx.Commit()
	}

	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("unable to complete package installation: %w", err)
	}

	return db, nil
}

func (p *Project) Migrate(dsn string) error {
	db, err := p.Open(dsn)
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return fmt.Errorf("unable to close database after migration: %w", err)
	}

	return nil
}

// Load all the schemas for all the packages from the project.
// This is likely to be expensive because it requires parsing the entire
// set of files in each package.
func (p *Project) parseSchemas() error {

	for _, pkg := range p.pkgs {
		err := pkg.readSchema()
		if err != nil {
			return err
		}
	}

	return nil
}

//go:embed pgpkg
var pgpkgFS embed.FS

// PGKSchemaName is the name of the pgpkg schema itself.
const PGKSchemaName = "pgpkg"

// Init initialises the pgpkg schema itself. It effectively uses pgpkg's
// migration tools to bookstrap itself.
func (p *Project) Init(tx *PkgTx) error {
	var isInitialised int
	err := tx.QueryRow("select count(*) from information_schema.schemata where schema_name = 'pgpkg'").Scan(&isInitialised)
	if err != nil {
		return fmt.Errorf("unable to read schema: %w", err)
	}

	basePkg, ok := p.pkgs["github.com/pgpkg/pgpkg"]
	if !ok {
		return fmt.Errorf("base package (github.com/pgpkg/pgpkg) not found")
	}

	if basePkg.SchemaNames[0] != PGKSchemaName {
		return fmt.Errorf("expected root schema name %s, got %s", PGKSchemaName, basePkg.SchemaNames[0])
	}

	// We can force the package to run all the migration scripts without
	// checking if they have been already run. This prevents the migration
	// from trying to lookup database tables before they are created.
	if isInitialised == 0 {
		basePkg.bootstrapSchema = true
	}

	return nil
}

// NewProject creates a new project. It adds the "pgpkg" package to the project, which is
// required to track the objects we create and remove.
func NewProject() *Project {
	p := &Project{
		pkgs: make(map[string]*Package),
	}

	// Always include the embedded pgpkg schema. This is treated specially, and is always
	// installed first.
	if _, err := p.AddEmbeddedFS(pgpkgFS, "pgpkg"); err != nil {
		panic(err)
	}

	return p
}

// NewProjectFrom creates a new project and adds the package found at path.
// It also configures (and possibly creates) a project cache, also rooted at the given path.
// If searchCaches is not nil, these will be searched in order when resolving dependencies.
// Search caches take precedence over the project cache.
func NewProjectFrom(pkgPath string, searchCaches ...Cache) (*Project, error) {
	p := NewProject()
	src, err := NewSource(pkgPath)
	if err != nil {
		return nil, err
	}

	if root, err := p.AddSource(src); err != nil {
		return nil, err
	} else {
		p.Root = root
	}

	// Get a cache from the top-level project source. If it's a writable cache, use it as the project
	// cache. Otherwise, add it as the top-level search cache.
	cache, err := src.Cache()
	if err != nil {
		fmt.Printf("warning: %v\n", err)
	} else if writeCache, ok := cache.(*WriteCache); ok {
		p.Cache = writeCache
	} else if cache != nil {
		p.Search = append(p.Search, cache)
	}

	p.Search = append(p.Search, searchCaches...)

	return p, nil
}

func (p *Project) PrintInfo(w InfoWriter) {
	var srcLocations []string
	for _, s := range p.Sources {
		srcLocations = append(srcLocations, s.Location())
	}

	w.Print("Project Sources", srcLocations)

	for name, pkg := range p.pkgs {
		w.Println("Package", name)
		pkg.PrintInfo(w)
		w.Println()
	}
}

// Parse prepares a project for migrating or other processing by resolving any dependencies
// and parsing the schemas.
func (p *Project) Parse() error {
	if err := p.resolveDependencies(); err != nil {
		return err
	}

	if err := p.parseSchemas(); err != nil {
		return err
	}

	return nil
}
