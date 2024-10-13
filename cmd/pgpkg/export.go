package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"os"
	"path"
)

func doExport() {
	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet("export", flag.ExitOnError)
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

	zipName := path.Base(p.Root.Name) + ".zip"

	zipFile, err := os.Create(zipName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pgpkg: unable to open ZIP file %s for writing: %v\n", zipName, err)
		os.Exit(1)
	}
	zipWriter := zip.NewWriter(zipFile)

	err = pgpkg.WriteProject(zipWriter, p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pgpkg: unable to export project: %v", err)
		os.Exit(1)
	}

	if err := zipWriter.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "pgpkg: unable to export project: %v", err)
		os.Exit(1)
	}

	fmt.Println("exported to", zipName)
}
