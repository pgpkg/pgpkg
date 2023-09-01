package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: pgpkg {deploy | repl | try | export | import} [options]")
}

// Search from the current directory backwards until we find a "pgpkg.toml" file,
// then return the directory in which it was found.
func findDefaultPkg() (string, error) {
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
		// This check is done last so that the current directory is always
		// searched, even if it's not inside the user's home.
		if !strings.HasPrefix(cwd, homeDir) {
			return "", fmt.Errorf("no package found")
		}
	}
}

// If args contains a package, return that. Otherwise, search for a target package.
func findPkg(args []string) (string, error) {
	if len(args) == 0 {
		return findDefaultPkg()
	}

	if len(args) == 1 {
		return args[0], nil
	}

	return "", fmt.Errorf("multiple package paths specified")
}
