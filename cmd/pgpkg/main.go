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
func createTempDB() (string, error) {
	dsn := os.Getenv("DSN")
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

// This command-line version of pgpkg takes one or more directories or ZIP files, and installs them into the database.
func main() {

	var replDb string
	var replDSN string

	// Don't commit to the database by accident.
	pgpkg.Options.DryRun = true

	// Keep a copy of the original DSN, since we modify it during REPL.
	defaultDSN := os.Getenv("DSN")

	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	repl := flag.Bool("repl", false, "Run the migration in a temp database, then start psql")

	flag.Parse()
	pkgPaths := flag.Args()
	if len(pkgPaths) < 1 {
		fmt.Fprintln(os.Stderr, "pgpkg: requires one or more packages to install or upgrade")
		os.Exit(1)
	}

	if *repl {
		var err error
		replDb, err = createTempDB()
		if err != nil {
			fmt.Fprintln(os.Stderr, "pgpkg: unable to create REPL database: %w", err)
			os.Exit(1)
		}

		// Add the REPL dbname to the DSN, which will override the PGDATABASE environment variable.
		// If there are two dbnames, only the last one is used, effectively overriding
		// anything in the environment.
		replDSN = os.Getenv("DSN")
		replDSN = replDSN + " dbname=" + replDb
		os.Setenv("DSN", replDSN)

		pgpkg.Options.DryRun = false
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

	if replDb != "" {
		err = doReplSession(replDSN)
		if err != nil {
			fmt.Fprintf(os.Stderr, "psql error: %v\n", err)
		}

		if err = dropTempDB(defaultDSN, replDb); err != nil {
			fmt.Fprintf(os.Stderr, "unable to drop REPL database %s: %v\n", replDb, err)
			os.Exit(1)
		}
	}
}
