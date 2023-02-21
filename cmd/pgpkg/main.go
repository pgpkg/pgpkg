package main

import (
	"flag"
	"fmt"
	"github.com/commandquery/pgpkg"
	"os"
)

func main() {
	// This simple version of pgpkg takes a single argument and installs it into the database.

	options := &pgpkg.Options{}
	flag.BoolVar(&options.Verbose, "verbose", false, "Print more information about what pgpkg has done")

	// Take the argument and look for a "pgpkg" directory under it.
	// (is this necessary? It doesn't seem like it should be)
	flag.Parse()
	pkgDir := flag.Arg(0)
	if pkgDir == "" {
		fmt.Fprintln(os.Stderr, "pgpkg: requires a package to install or upgrade")
		os.Exit(1)
	}

	pkgFS := os.DirFS(pkgDir)
	if pkgFS == nil {
		panic(fmt.Errorf("unable to open package %s", pkgDir))
	}

	// Load the package. LoadPackage currently expects files in ./api, ./schema and ./tests.
	pkg, err := pgpkg.LoadPackage("embedded", pkgFS, options)
	if err != nil {
		panic(err)
	}

	db, err := pgpkg.Open("", options)
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	// Initialise pgpkg itself.
	if err = pgpkg.Init(tx, options); err != nil {
		panic(err)
	}

	// Apply the requested package.
	if err = pkg.Apply(tx); err != nil {
		panic(err)
	}

	err = tx.Commit()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if options.Verbose {
		fmt.Printf("installed %d function(s), %d view(s) and %d trigger(s). %d migration(s) needed.\n",
			pkg.StatFuncCount, pkg.StatViewCount, pkg.StatTriggerCount, pkg.StatMigrationCount)
	}
}
