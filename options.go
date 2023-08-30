package pgpkg

import (
	"fmt"
	"os"
	"regexp"
)

// Options is a list of global options used by pgpkg.

var Options struct {
	Verbose        bool           // print lots of stuff
	Summary        bool           // print a summary of the installation
	DryRun         bool           // rollback after installation (default)
	ShowTests      bool           // Show the result of each SQL test that was run.
	SkipTests      bool           // Don't run the tests. Useful when fixing them!
	IncludePattern *regexp.Regexp // Pattern to use for running tests
	ExcludePattern *regexp.Regexp // Pattern to use for running tests
}

func showHelp() {
	fmt.Println(`pgpkg - postgresql packaging and migration tool.

Usage

    pgpkg [options] pkg

where "pkg" is a directory (or the child of a directory) containing "pgpkg.toml", or a ZIP file.

pgpkg works using the current Postgresql environment (PGHOST, PGDATABASE, ...).

Transactions and Testing Options

--dry-run
    perform a full schema migration (including tests), but don't commit the results.
    The database is therefore left unchanged.

--show-tests
    This option prints a pass/fail status for each test that's run.

--skip-tests
    do not run any tests before committing the changes. You should take care with this option.

--include-tests=[regexp]
    only run tests whose SQL function name matches the given regexp.

--exclude-tests=[regexp]
    run all tests, except those whose SQL function name matches the given regexp.

Logging Options

pgpkg normally runs silently (unless your SQL code includes raise notice messages). These options tell pgpkg
to display more information:

--verbose
    This option print logs describing what pgpkg is up to.

--summary
    This option print a summary of pgpkg operations when it finishes.`)
}

// ParseArgs parses the os.Args for a standard set of OS arguments.
// ParseArgs deletes matching arguments from os.Args so that the caller
// doesn't need to worry about them.
//
// When embedding pgpkg into your own programs, set prefix to "pgpkg" to
// differentiate pgpkg arguments from your own. Doing this will make it possible
// to set pgpkg options from your code are runtime with prefixed options such as "--dry-run".
// If "prefix" is empty ("") then options will not be prefixed; ie, "--dry-run".
//
// You should call ParseArgs before calling flag.Parse() if you are using the
// standard flag library.
func ParseArgs(prefix string) error {

	if prefix != "" {
		prefix = prefix + "-"
	}

	argPattern := fmt.Sprintf("^-?-%s([^=]+)($|=)", prefix)
	argExp := regexp.MustCompile(argPattern)

	var parsedArgs []string
	for _, a := range os.Args {
		pgpkgArgs := argExp.FindStringSubmatch(a)
		if pgpkgArgs == nil {
			parsedArgs = append(parsedArgs, a)
			continue
		}

		switchName := pgpkgArgs[1]

		switch switchName {
		case "verbose":
			Options.Verbose = true
			Options.ShowTests = true
			Options.Summary = true

		case "summary":
			Options.Summary = true

		case "show-tests":
			Options.ShowTests = true

		case "skip-tests":
			Options.SkipTests = true

		case "include-tests":
			var err error
			Options.IncludePattern, err = regexp.Compile(a[22:]) // full argument is --include-tests=
			if err != nil {
				return fmt.Errorf("unable to compile pattern %s", a[14:])
			}

		case "exclude-tests":
			var err error
			Options.ExcludePattern, err = regexp.Compile(a[22:]) // full argument is --include-tests=
			if err != nil {
				return fmt.Errorf("unable to compile pattern %s", a[14:])
			}

		case "help":
			showHelp()
			return ErrUserRequest

		// We're not the argument police. We're just here to look for
		// the args we know about.
		default:
			parsedArgs = append(parsedArgs, a)
		}
	}

	// Remove any pgpkg args from the os list so that the flag package
	// can deal with app-specific args.
	os.Args = parsedArgs
	return nil
}
