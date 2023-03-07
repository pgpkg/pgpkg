package main

import (
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"os"
)

// This command-line version of pgpkg takes one or more directories or ZIP files, and installs them into the database.
func main() {

	options := &pgpkg.Options{}
	flag.BoolVar(&options.Verbose, "verbose", false, "Print lots of information about what pgpkg is doing")
	flag.BoolVar(&options.Summary, "summary", false, "Print a summary of the packages installed/updated")

	// Take the argument and look for a "pgpkg" directory under it.
	// (is this necessary? It doesn't seem like it should be)
	flag.Parse()
	pkgPaths := flag.Args()
	if len(pkgPaths) < 1 {
		fmt.Fprintln(os.Stderr, "pgpkg: requires one or more packages to install or upgrade")
		os.Exit(1)
	}

	var p pgpkg.Project
	p.AddPath(pkgPaths...)

	db, err := p.Open("", options)
	if err != nil {
		pgpkg.Exit(err)
	}

	err = db.Close()
	if err != nil {
		pgpkg.Exit(err)
	}
}
