package main

import (
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"os"
)

func doTest(dsn string) {
	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	// This is set by default with pgpkg test.
	pgpkg.Options.ShowTests = true

	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet("test", flag.ExitOnError)
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		pgpkg.Exit(fmt.Errorf("unable to parse arguments: %w", err))
	}

	// The purpose of "pgpkg test" is just to build the schema in a test database
	// and return, reporting any errors along the way. So that's what we do!
	tempDB, err := initTempDb(dsn, flagSet)
	if err != nil {
		pgpkg.Exit(err)
	}

	dropTempDBOrExit(dsn, tempDB.DBName)
}
