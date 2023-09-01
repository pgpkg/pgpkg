package main

import (
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"os"
)

func doImport() {
	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet("import", flag.ExitOnError)
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		pgpkg.Exit(fmt.Errorf("unable to parse arguments: %w", err))
	}

	// unlike other commands, pgpkg import can have two positional parameters, being the target package
	// and the package being imported (source package).
	args := flagSet.Args()

	// We want to import srcPkgPath into targetPkgPath
	var targetPkgPath, srcPkgPath string
	var err error

	switch len(args) {
	case 1:
		targetPkgPath, err = findDefaultPkg()
		srcPkgPath = args[0]
	case 2:
		targetPkgPath = args[0]
		srcPkgPath = args[1]
	default:
		pgpkg.Exit(fmt.Errorf("usage: pgpkg import [target] <source>"))
	}

	p, err := pgpkg.NewProjectFrom(targetPkgPath)
	if err != nil {
		pgpkg.Exit(err)
	}

	if p.Cache == nil {
		pgpkg.Exit(fmt.Errorf("project has no cache"))
	}

	// Load the project which is to be imported. Dependencies are resolved using the
	// targe project cache first. This means that if a dependency is already imported,
	// there won't be an error, even if the source package doesn't have the dependency cached.
	i, err := pgpkg.NewProjectFrom(srcPkgPath, &p.Cache.ReadCache)
	if err != nil {
		pgpkg.Exit(err)
	}

	if i.Root.Name == p.Root.Name {
		pgpkg.Exit(fmt.Errorf("cowardly refusing to import a project into itself"))
	}

	if err := p.Cache.ImportProject(i); err != nil {
		pgpkg.Exit(err)
	}

	if p.Root.AddUses(i.Root.Name) {
		// Uses clause added, need to write the config out
		if err := p.Root.WriteConfig(); err != nil {
			pgpkg.Exit(fmt.Errorf("unable to write package config: %w", err))
		}
	}
}
