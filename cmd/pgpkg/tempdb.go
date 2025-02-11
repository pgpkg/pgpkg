package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
)

type TempDB struct {
	DSN     string
	DBName  string
	PkgPath string
	Project *pgpkg.Project
}

// Set up a project in a temp DB, and return the database's name.
// Before exiting, the database should be removed by the caller with dropTempDBOrExit().
// This is used by "pgpkg repl" and "pgpgk test".
func initTempDb(dsn string, flagSet *flag.FlagSet) (*TempDB, error) {
	pkgPath, err := findPkg(flagSet.Args())
	if err != nil {
		return nil, err
	}

	p, err := pgpkg.NewProjectFrom(pkgPath)
	if err != nil {
		return nil, err
	}

	tempDbName, err := pgpkg.CreateTempDB(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgpkg: unable to create REPL database: %w\n", err)
	}

	// Add the REPL dbname to the DSN, which will override the PGDATABASE environment variable.
	// If there are two dbnames, only the last one is used, effectively overriding
	// anything in the environment.
	tempDSN := dsn + " dbname=" + tempDbName

	pgpkg.Options.DryRun = false

	err = p.Migrate(tempDSN)
	if err != nil {
		// Clean up the database if there's an error; the caller will probably forget to do so.
		dropErr := pgpkg.DropTempDB(dsn, tempDbName)
		return nil, errors.Join(err, dropErr)
	}

	return &TempDB{
		DSN:     tempDSN,
		DBName:  tempDbName,
		PkgPath: pkgPath,
		Project: p,
	}, nil
}
