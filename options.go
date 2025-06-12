package pgpkg

import (
	"fmt"
	"os"
	"regexp"
)

// Options is a list of global options used by pgpkg.

var Options struct {
	Verbose         bool           // print lots of stuff
	Summary         bool           // print a summary of the installation
	DryRun          bool           // rollback after installation (default)
	ShowTests       bool           // Show the result of each SQL test that was run.
	SortTests       bool           // Execute tests in a well defined order
	ShowSkipped     bool           // Show skipped tests
	SkipTests       bool           // Don't run the tests. Useful when fixing them!
	KeepTestScripts bool           // Keep the test functions, useful for Go unit testing, use only with temporary databases.
	IncludePattern  *regexp.Regexp // Pattern to use for running tests
	ExcludePattern  *regexp.Regexp // Pattern to use for running tests
	ForceRole       string         // Use this role instead of package roles
}

func showHelp() {
	fmt.Println(`pgpkg - postgresql packaging and migration tool.

Usage

    pgpkg [options] [pkg]

where "pkg" is a directory (or the child of a directory) containing "pgpkg.toml", or a ZIP file.
If not set, pkg searches for pgpkg.toml in the current and parent directories.

pgpkg works using the current Postgresql environment (PGHOST, PGDATABASE, ...).

Transactions and Testing Options

--dry-run
    Perform a full schema migration (including tests), but don't commit the results.
    The database is therefore left unchanged.

--show-tests
    This option prints a pass/fail status for each test that's run.

--sort-tests
    Tests normally run in random order. This option runs tests in a fixed order,
	so that the test run is repeatable. This can be helpful during large refactors.

--skip-tests
    Do not run any tests before committing the changes. You should take care with this option.

--keep-test-scripts
    Do not purge the test scripts when finished. Useful for integrating with Go unit tests.
	WARNING: Migrations performed using -keep-test-scripts cannot be upgraded later.
	Only use this option on temporary, disposable databases. 

--include-tests=[regexp]
    Only run tests whose SQL function name matches the given regexp.

--exclude-tests=[regexp]
    Run all tests, except those whose SQL function name matches the given regexp.

--show-skipped
    Logs all tests, even if they are skipped. By default, only tests that run are logged.

Logging Options

pgpkg normally runs silently (unless your SQL code includes raise notice messages). These options tell pgpkg
to display more information:

--verbose
    This option prints logs describing what pgpkg is up to.

--summary
    This option prints a summary of pgpkg operations when it finishes.`)
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

	argPattern := fmt.Sprintf("^-?-%s([^=]+)(=(.*))?$", prefix)
	argExp := regexp.MustCompile(argPattern)

	var parsedArgs []string
	for _, a := range os.Args {
		pgpkgArgs := argExp.FindStringSubmatch(a)
		if pgpkgArgs == nil {
			parsedArgs = append(parsedArgs, a)
			continue
		}

		switchName := pgpkgArgs[1]
		switchValue := pgpkgArgs[3]

		switch switchName {
		case "argtest":
			fmt.Println(switchValue)
			os.Exit(1)

		case "verbose":
			Options.Verbose = true
			Options.ShowTests = true
			Options.Summary = true

		case "summary":
			Options.Summary = true

		case "show-tests":
			Options.ShowTests = true

		case "sort-tests":
			Options.SortTests = true

		case "show-skipped":
			Options.ShowSkipped = true

		case "skip-tests":
			Options.SkipTests = true

		case "keep-test-scripts":
			Options.KeepTestScripts = true

		case "include-tests":
			var err error
			Options.IncludePattern, err = regexp.Compile(switchValue)
			if err != nil {
				return fmt.Errorf("unable to compile pattern %s", switchValue)
			}

		case "exclude-tests":
			var err error
			Options.ExcludePattern, err = regexp.Compile(switchValue)
			if err != nil {
				return fmt.Errorf("unable to compile pattern %s", switchValue)
			}

		case "force-role":
			Options.ForceRole = switchValue

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
