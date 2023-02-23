package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/commandquery/pgpkg"
	"os"
)

func printError(err error) {

	var pkgErr *pgpkg.PKGError
	ok := errors.As(err, &pkgErr)
	if !ok {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}

	pkgErr.PrintRootContext(2)

	//fmt.Println("found pkgerror:", pkgErr)
	//fmt.Println("found context:", pkgErr.Context)

	//root := pkgErr.Root()
	//fmt.Println("found root:", root)
	//fmt.Println("root context:", root.GetContext())
}

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
		fmt.Fprintf(os.Stderr, "unable to open package %s", pkgDir)
		os.Exit(1)
	}

	// Load the package. LoadPackage currently expects files in ./api, ./schema and ./tests.
	pkg, err := pgpkg.LoadPackage(pkgDir, pkgFS, options)
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	db, err := pgpkg.Open("", options)
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	tx, err := db.Begin()
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	// Initialise pgpkg itself.
	if err = pgpkg.Init(tx, options); err != nil {
		printError(err)
		os.Exit(1)
	}

	// Apply the requested package.
	if err = pkg.Apply(tx); err != nil {
		printError(err)
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	if options.Verbose {
		fmt.Printf("installed %d function(s), %d view(s) and %d trigger(s). %d migration(s) needed.\n",
			pkg.StatFuncCount, pkg.StatViewCount, pkg.StatTriggerCount, pkg.StatMigrationCount)
	}
}
