package main

import (
	"archive/zip"
	"database/sql"
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
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

func deploy(commit bool) {
	pgpkg.Options.DryRun = !commit

	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	pkgPath, err := findPkgPath()
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

func dropTempDBOrExit(dsn string, replDb string) {
	if err := dropTempDB(dsn, replDb); err != nil {
		fmt.Fprintf(os.Stderr, "unable to drop REPL database %s: %v\n", replDb, err)
		os.Exit(1)
	}
}

func repl() {
	var replDb string
	var replDSN string

	// Keep a copy of the original DSN, since we modify it during REPL.
	defaultDSN := os.Getenv("DSN")

	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	var pkgPath string
	var err error

	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet("repl", flag.ExitOnError)
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		pgpkg.Exit(fmt.Errorf("unable to parse arguments: %w", err))
	}

	if flagSet.NArg() == 0 {
		// No args: find project in current or parent dir
		pkgPath, err = findPkgPath()
		if err != nil {
			pgpkg.Exit(err)
		}
	} else if flagSet.NArg() == 1 {
		// single arg: find project in specific directory.
		pkgPath = flagSet.Arg(0)
	} else {
		// multiple args is an error
		usage()
		os.Exit(1)
	}

	p, err := pgpkg.NewProjectFrom(pkgPath)
	if err != nil {
		pgpkg.Exit(err)
	}

	replDb, err = createTempDB()
	if err != nil {
		pgpkg.Exit(fmt.Errorf("pgpkg: unable to create REPL database: %w\n", err))
	}

	defer dropTempDBOrExit(defaultDSN, replDb)

	// Add the REPL dbname to the DSN, which will override the PGDATABASE environment variable.
	// If there are two dbnames, only the last one is used, effectively overriding
	// anything in the environment.
	replDSN = os.Getenv("DSN")
	replDSN = replDSN + " dbname=" + replDb
	os.Setenv("DSN", replDSN)

	pgpkg.Options.DryRun = false

	err = p.Migrate()
	if err != nil {
		pgpkg.Exit(err)
	}

	err = doReplSession(replDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "psql error: %v\n", err)
	}
}

func export() {
	pkgPath, err := findPkgPath()
	if err != nil {
		pgpkg.Exit(err)
	}

	p, err := pgpkg.NewProjectFrom(pkgPath)
	if err != nil {
		pgpkg.Exit(err)
	}

	// in-memory zip
	zipName := "pgpkg.zip"
	zipFile, err := os.Create(zipName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pgpkg: unable to open ZIP file %s for writing: %v\n", zipName, err)
		os.Exit(1)
	}
	zipWriter := zip.NewWriter(zipFile)

	err = pgpkg.WriteProject(zipWriter, p)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pgpkg: unable to export project: %v", err)
		os.Exit(1)
	}

	if err := zipWriter.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "pgpkg: unable to export project: %v", err)
		os.Exit(1)
	}

	fmt.Println("exported to", zipName)
}

func importPackage() {
	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	pkgPath, err := findPkgPath()
	if err != nil {
		pgpkg.Exit(err)
	}

	p, err := pgpkg.NewProjectFrom(pkgPath)
	if err != nil {
		pgpkg.Exit(err)
	}

	if p.Cache == nil {
		pgpkg.Exit(fmt.Errorf("project has no cache"))
	}

	// os.Args[0]=program name, os.Args[1]=command verb, os.Args[2]=import filename(s)
	if len(os.Args) != 3 {
		usage()
		fmt.Println("pgpkg import requires a package path to import")
	}

	importPkgPath := os.Args[2]
	// Load the project which is to be imported. Dependencies are resolved using the
	// targe project cache first. This means that if a dependency is already imported,
	// there won't be an error, even if the source package doesn't have the dependency cached.
	i, err := pgpkg.NewProjectFrom(importPkgPath, &p.Cache.ReadCache)
	if err != nil {
		pgpkg.Exit(err)
	}

	if i.Root.Name == p.Root.Name {
		pgpkg.Exit(fmt.Errorf("cowardly refusing to import a project into itself"))
	}

	if err := p.Cache.ImportProject(i); err != nil {
		pgpkg.Exit(err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: pgpkg {deploy | repl | try | export | import} [options]")
}

// Search from the current directory backwards until we find a "pgpkg.toml" file,
// then return the directory in which it was found.
func findPkgPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	for {
		_, err := os.Stat(path.Join(cwd, "pgpkg.toml"))
		if err == nil {
			return cwd, nil
		}

		// Get the parent of the current working directory
		cwd = path.Dir(cwd)

		// Only search until the home directory of the current user.
		if !strings.HasPrefix(cwd, homeDir) {
			return "", fmt.Errorf("no package found")
		}
	}
}

// This command-line version of pgpkg takes one or more directories or ZIP files, and installs them into the database.
func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "deploy":
		deploy(true)

	case "try":
		deploy(false)

	case "repl":
		repl()

	case "export":
		export()

	case "import":
		importPackage()

	default:
		usage()
		os.Exit(1)
	}
}
