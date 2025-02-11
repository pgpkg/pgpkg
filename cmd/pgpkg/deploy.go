package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/pgpkg/pgpkg"
	"github.com/rjeczalik/notify"
	"os"
)

// Start the watch process. Returns an error if the watch can't start.
// This function never returns; the user is expected to hit ctrl-c
func doDeployWatch(pkgPath string, dsn string, prompt bool, exitChan chan error) {
	watch, err := NewWatch(pkgPath)
	if err != nil {
		exitChan <- err
		return
	}

	watch.Watch(func(e []notify.EventInfo) {
		if prompt {
			fmt.Print("\r")
			pgpkg.Stdout.Printf("[%s] %s\n", "watch", "updating schema")
		}

		p, err := pgpkg.NewProjectFrom(pkgPath)
		if err == nil {
			err = p.Migrate(dsn)
		}

		if err != nil && !errors.Is(err, pgpkg.ErrUserRequest) {
			pgpkg.Stdout.Printf("[%s] %s\n", "watch", err)
		} else {
			pgpkg.Stdout.Printf("[%s] %s\n", "watch", "updated successfully")
		}

		if prompt {
			fmt.Print("pgpkg> ")
		}
	})
}

func apply(subcommand string, dsn string, commit bool) {
	pgpkg.Options.DryRun = !commit

	if err := pgpkg.ParseArgs(""); err != nil {
		pgpkg.Exit(err)
	}

	// This is here just so we can easily add new flags later if needed.
	flagSet := flag.NewFlagSet(subcommand, flag.ExitOnError)
	watchFlag := flagSet.Bool("watch", false, "(experimental) do not exit; watch for changes and reload them")
	replFlag := flagSet.Bool("repl", false, "start a REPL session after deployment (use with --watch)")
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

	err = p.Migrate(dsn)

	delayedExit := *replFlag || *watchFlag

	// We don't want to terminate if we're in watch mode.
	if !delayedExit || (err != nil && !errors.Is(err, pgpkg.ErrUserRequest)) {
		pgpkg.Exit(err)
	}

	// Either watch or repl needs to write to the exitChan when terminating.
	exitChan := make(chan error)

	if *watchFlag {
		pgpkg.Stdout.Println("[warning] --watch is experimental; use with care. please report issues to https://github.com/pgpkg/pgpkg/issues")
		go doDeployWatch(pkgPath, dsn, *replFlag, exitChan)
	}

	if *replFlag {
		go func() {
			exitChan <- doReplSession(dsn)
		}()
	}

	exitErr := <-exitChan
	pgpkg.Exit(exitErr)
}

func doDeploy(dsn string) {
	apply("deploy", dsn, true)
}
