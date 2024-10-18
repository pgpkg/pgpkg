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
func doWatch(pkgPath string, dsn string) {
	watch, err := NewWatch(pkgPath)
	if err != nil {
		pgpkg.Exit(err)
	}

	watch.Watch(func(e []notify.EventInfo) {
		p, err := pgpkg.NewProjectFrom(pkgPath)
		if err == nil {
			err = p.Migrate(dsn)
		}

		if err != nil && !errors.Is(err, pgpkg.ErrUserRequest) {
			pgpkg.Stdout.Printf("[%s] %s\n", "watch", err)
		} else {
			pgpkg.Stdout.Printf("[%s] %s\n", "watch", "updated successfully")
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
	watchFlag := flagSet.Bool("watch", false, "do not quit; watch for changes and reload them")
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

	// We don't want to terminate if we're in watch mode.
	if !*watchFlag || (err != nil && !errors.Is(err, pgpkg.ErrUserRequest)) {
		pgpkg.Exit(err)
	}

	doWatch(pkgPath, dsn)
}

func doDeploy(dsn string) {
	apply("deploy", dsn, true)
}
