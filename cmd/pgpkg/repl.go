package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
)

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

// Start an interactive psql session with the given DSN, and wait for it to exit.
func doReplSession(dsn string) error {
	// Create a new command
	cmd := exec.Command("psql", "-v", "PROMPT1=pgpkg> ", "-v", "PROMPT2=pgpkg| ", dsn)

	// Set the command's input and output to standard input and output
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start psql: %w", err)
	}

	signal.Ignore(os.Interrupt)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("psql completed with an error: %w", err)
	}

	signal.Reset(os.Interrupt)

	return nil
}

func dropTempDBOrExit(dsn string, replDb string) {
	if err := dropTempDB(dsn, replDb); err != nil {
		fmt.Fprintf(os.Stderr, "unable to drop REPL database %s: %v\n", replDb, err)
		os.Exit(1)
	}
}

func doRepl(dsn string) {
	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet("repl", flag.ExitOnError)
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		pgpkg.Exit(fmt.Errorf("unable to parse arguments: %w", err))
	}

	tempDbName, err := initTempDb(dsn, flagSet)
	if err != nil {
		pgpkg.Exit(err)
	}

	defer dropTempDBOrExit(dsn, tempDbName)

	if err = doReplSession(tempDbName); err != nil {
		fmt.Fprintf(os.Stderr, "psql error: %v\n", err)
	}
}

func doTest(dsn string) {
	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet("test", flag.ExitOnError)
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		pgpkg.Exit(fmt.Errorf("unable to parse arguments: %w", err))
	}

	// This is set by default with pgpkg test.
	pgpkg.Options.ShowTests = true

	// The purpose of "pgpkg test" is just to build the schema in a test database
	// and return, reporting any errors along the way. So that's what we do!
	tempDbName, err := initTempDb(dsn, flagSet)
	if err != nil {
		pgpkg.Exit(err)
	}

	dropTempDBOrExit(dsn, tempDbName)
}

// Set up the project in a temp DB, and return the database's name.
// Before exiting, the database should be removed by the caller with dropTempDBOrExit().
// This is used by "pgpkg repl" and "pgpgk test".
func initTempDb(dsn string, flagSet *flag.FlagSet) (string, error) {
	pkgPath, err := findPkg(flagSet.Args())
	if err != nil {
		return "", err
	}

	p, err := pgpkg.NewProjectFrom(pkgPath)
	if err != nil {
		return "", err
	}

	tempDbName, err := createTempDB(dsn)
	if err != nil {
		return "", fmt.Errorf("pgpkg: unable to create REPL database: %w\n", err)
	}

	// Add the REPL dbname to the DSN, which will override the PGDATABASE environment variable.
	// If there are two dbnames, only the last one is used, effectively overriding
	// anything in the environment.
	tempDSN := dsn + " dbname=" + tempDbName

	pgpkg.Options.DryRun = false

	err = p.Migrate(tempDSN)
	if err != nil {
		// Clean up the database if there's an error; the caller will probably forget to do so.
		dropTempDBOrExit(dsn, tempDSN)
		return "", err
	}

	return tempDbName, nil
}
