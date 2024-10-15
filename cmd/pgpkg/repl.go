package main

import (
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"github.com/rjeczalik/notify"
	"os"
	"os/exec"
	"os/signal"
)

// Start an interactive psql session with the given DSN, and wait for it to exit.
func doReplSession(tempDB *TempDB) error {
	// Create a new command
	cmd := exec.Command("psql", "-v", "PROMPT1=pgpkg> ", "-v", "PROMPT2=pgpkg| ", tempDB.DSN)

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

func doRepl(dsn string) {
	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet("repl", flag.ExitOnError)
	watchFlag := flagSet.Bool("watch", false, "watch for changes")
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		pgpkg.Exit(fmt.Errorf("unable to parse arguments: %w", err))
	}

	tempDB, err := initTempDb(dsn, flagSet)
	if err != nil {
		pgpkg.Exit(err)
	}

	defer dropTempDBOrExit(dsn, tempDB.DBName)

	if *watchFlag {
		go watch(tempDB.Path, func(e notify.EventInfo) {
			fmt.Println("received watch event:", e)
		})
	}
	if err = doReplSession(tempDB); err != nil {
		fmt.Fprintf(os.Stderr, "psql error: %v\n", err)
	}
}
