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
	DryRun         bool           // rollback after installation (good for testing)
	ShowTests      bool           // Show the result of each SQL test that was run.
	SkipTests      bool           // Don't run the tests. Useful when fixing them!
	IncludePattern *regexp.Regexp // Pattern to use for running tests
	ExcludePattern *regexp.Regexp // Pattern to use for running tests
}

var argPattern = regexp.MustCompile("^-?-pgpkg-([^=]+)($|=)")

// ParseArgs parses the os.Args for the standard set of --pgpkg-<arg>s.
// ParseArgs deletes matching arguments from os.Args so that the caller
// doesn't need to worry about them.
//
// You should call ParseArgs before calling flag.Parse() if you are using the
// standard flag library.
func ParseArgs() error {
	var parsedArgs []string
	for _, a := range os.Args {
		pgpkgArgs := argPattern.FindStringSubmatch(a)
		if pgpkgArgs == nil {
			parsedArgs = append(parsedArgs, a)
			continue
		}

		switchName := pgpkgArgs[1]

		switch switchName {
		case "dry-run":
			Options.DryRun = true

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
			Options.IncludePattern, err = regexp.Compile(a[22:]) // full argument is --pgpkg-include-tests=
			if err != nil {
				return fmt.Errorf("unable to compile pattern %s", a[14:])
			}

		case "exclude-tests":
			var err error
			Options.ExcludePattern, err = regexp.Compile(a[22:]) // full argument is --pgpkg-include-tests=
			if err != nil {
				return fmt.Errorf("unable to compile pattern %s", a[14:])
			}

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
