package pgpkg

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
)

// These functions create and destroy tempoarary databases.
// They are used in pgpkg repl and pgpkg test, but are also
// useful when writing unit tests in Go.

// MkTempDbName returns a SAFE, random, database name fragment.
// Take care to ensure that any changes to this function return names that are always safe to use
// in un-escaped SQL statements.
func MkTempDbName() string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// CreateTempDB creats a temporary database with a random name.
// We do this by connecting to the database using the environment,
// and running "create database". We return the database name.
func CreateTempDB(dsn string) (string, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return "", fmt.Errorf("unable to open database: %w", err)
	}

	// important: ensure that the dbname only ever contains alphanumeric characters
	dbname := "pgpkg." + MkTempDbName()
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

// DropTempDB drops the given database. WARNING: it will actually drop any database
// you ask it to, so take care only to use the database created by CreateTempDb
func DropTempDB(dsn string, dbname string) error {
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

func DropTempDBOrExit(dsn string, replDb string) {
	if err := DropTempDB(dsn, replDb); err != nil {
		fmt.Fprintf(os.Stderr, "unable to drop REPL database %s: %v\n", replDb, err)
		os.Exit(1)
	}
}
