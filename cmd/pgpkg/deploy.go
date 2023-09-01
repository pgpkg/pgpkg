package main

import (
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"os"
)

func apply(commit bool) {
	pgpkg.Options.DryRun = !commit

	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet("deploy", flag.ExitOnError)
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		pgpkg.Exit(fmt.Errorf("unable to parse arguments: %w", err))
	}

	pkgPath, err := findPkg(flagSet.Args())
	if err != nil {
		pgpkg.Exit(err)
	}

	p, err := pgpkg.NewProjectFrom(pkgPath)
	if err != nil {
		pgpkg.Exit(err)
	}

	err = p.Migrate()
	if err != nil {
		pgpkg.Exit(err)
	}
}

func doDeploy() {
	apply(true)
}