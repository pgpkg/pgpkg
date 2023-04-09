package main

import (
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"os"
)

// This command-line version of pgpkg takes one or more directories or ZIP files, and installs them into the database.
func main() {

	if err := pgpkg.ParseArgs(); err != nil {
		pgpkg.Exit(err)
	}

	flag.Parse()
	pkgPaths := flag.Args()
	if len(pkgPaths) < 1 {
		fmt.Fprintln(os.Stderr, "pgpkg: requires one or more packages to install or upgrade")
		os.Exit(1)
	}

	p := pgpkg.NewProject()
	p.AddPath(pkgPaths...)

	db, err := p.Open()
	if err != nil {
		pgpkg.Exit(err)
	}

	err = db.Close()
	if err != nil {
		pgpkg.Exit(err)
	}
}
