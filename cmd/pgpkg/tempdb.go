package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"math/rand"
	"os"
)

type TempDB struct {
	DSN     string
	DBName  string
	PkgPath string
	Project *pgpkg.Project
}

// Return a SAFE, random, database name fragment.
// Take care to ensure that any changes to this function return names that are always safe to use
// in un-escaped SQL statements.
func mkTempDbName() string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// Create a temporary database with a random name.
// We do this by connecting to the database using the environment,
// and running "create database".
func createTempDB(dsn string) (string, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return "", fmt.Errorf("unable to open database: %w", err)
	}

	// important: ensure that the dbname only ever contains alphanumeric characters
	dbname := "pgpkg." + mkTempDbName()
	mkdbcmd := fmt.Sprintf("create database \"%s\"", dbname)
	_, err = db.Exec(mkdbcmd)
	if err != nil {
		return "", fmt.Errorf("unable to create temp database \"%s\": %w", dbname, err)
	}

	if err := db.Close(); err != nil {
		return "", fmt.Errorf("unable to close database: %w", err)
	}

	return dbname, nil
}

func dropTempDB(dsn string, dbname string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("unable to open database: %w", err)
	}

	// important: ensure that the dbname only ever contains alphanumeric characters
	mkdbcmd := fmt.Sprintf("drop database \"%s\"", dbname)
	_, err = db.Exec(mkdbcmd)
	if err != nil {
		return fmt.Errorf("unable to drop temp database \"%s\": %w", dbname, err)
	}

	if err = db.Close(); err != nil {
		return fmt.Errorf("unable to close database: %w", err)
	}

	return nil
}

func dropTempDBOrExit(dsn string, replDb string) {
	if err := dropTempDB(dsn, replDb); err != nil {
		fmt.Fprintf(os.Stderr, "unable to drop REPL database %s: %v\n", replDb, err)
		os.Exit(1)
	}
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

	tempDbName, err := createTempDB(dsn)
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
		dropErr := dropTempDB(dsn, tempDbName)
		return nil, errors.Join(err, dropErr)
	}

	return &TempDB{
		DSN:     tempDSN,
		DBName:  tempDbName,
		PkgPath: pkgPath,
		Project: p,
	}, nil
}
