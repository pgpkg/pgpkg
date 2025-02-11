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

// Start the watch process. Returns an error if the watch can't start.
// Otherwise, a goroutine is started which performs the watch operation,
// and this function returns nil.
func startReplWatch(tempDB *TempDB) error {
	watch, err := NewWatch(tempDB.PkgPath)
	if err != nil {
		pgpkg.Exit(err)
	}

	go watch.Watch(func(e []notify.EventInfo) {
		fmt.Print("\r")
		p, err := pgpkg.NewProjectFrom(tempDB.PkgPath)
		if err == nil {
			err = p.Migrate(tempDB.DSN)
		}

		if err != nil {
			pgpkg.Stdout.Printf("[%s] %s\n", "watch", err)
		} else {
			pgpkg.Stdout.Printf("[%s] %s\n", "watch", "updated successfully")
		}
		fmt.Print("pgpkg> ")
	})

	return nil
}

func doRepl(dsn string) {
	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet("repl", flag.ExitOnError)
	watchFlag := flagSet.Bool("watch", false, "(experimental) watch for changes, and reload the schema as needed")
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		pgpkg.Exit(fmt.Errorf("unable to parse arguments: %w", err))
	}

	tempDB, err := initTempDb(dsn, flagSet)
	if err != nil {
		pgpkg.Exit(err)
	}

	defer pgpkg.DropTempDBOrExit(dsn, tempDB.DBName)

	if *watchFlag {
		if err = startReplWatch(tempDB); err != nil {
			pgpkg.Exit(err)
		}

		pgpkg.Stdout.Println("[warning] --watch is experimental; use with care. please report issues to https://github.com/pgpkg/pgpkg/issues")
	}

	if err = doReplSession(tempDB.DSN); err != nil {
		fmt.Fprintf(os.Stderr, "psql error: %v\n", err)
	}
}
