package pgpkg

import (
	"os"
	"regexp"
)

// Options is a list of global options used by pgpkg.

var Options struct {
	Verbose bool
	Summary bool
	DryRun  bool
}

var argPattern = regexp.MustCompile("^-?-pgpkg-(.+)$")

// ParseArgs parses the os.Args for the standard set of --pgpkg-<arg>s.
// ParseArgs deletes matching arguments from os.Args so that the caller
// doesn't need to worry about them.
//
// You should call ParseArgs before calling flag.Parse() if you are using the
// standard flag library.
func ParseArgs() {
	var parsedArgs []string
	for _, a := range os.Args {
		pgpkgArgs := argPattern.FindStringSubmatch(a)
		if pgpkgArgs == nil {
			parsedArgs = append(parsedArgs, a)
			continue
		}

		switch pgpkgArgs[1] {
		case "dry-run":
			Options.DryRun = true

		case "verbose":
			Options.Verbose = true

		case "summary":
			Options.Summary = true

		// We're not the argument police. We're just here to look for
		// the args we know about.
		default:
			parsedArgs = append(parsedArgs, a)
		}
	}

	// Remove any pgpkg args from the os list so that the flag package
	// can deal with app-specific args.
	os.Args = parsedArgs
}
