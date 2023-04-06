package pgpkg

import (
	"fmt"
	"regexp"
)

// These patterns are used to ensure that SQL can't be injected into statements.
// Take care to ensure that any changes can't be used against us.
var schemaPattern = regexp.MustCompile("^[a-z0-9][-_a-z0-9]*$")    // schema names
var rolePattern = regexp.MustCompile("^[$a-z0-9][-._/a-z0-9]*$")   // role names (can have leading $)
var extensionPattern = regexp.MustCompile("^[a-z0-9][-_a-z0-9]*$") // database extension names

// Sanitize checks that an identifier is valid per the given regexp, and panics if it doesn't.
// It's meant to be the last line of defence when SQL can't use interpolation.

func Sanitize(pattern *regexp.Regexp, v string) string {
	if pattern.MatchString(v) {
		return v
	}

	panic(fmt.Errorf("illegal identifier: %s", v))
}

func SanitizeSlice(pattern *regexp.Regexp, values []string) []string {
	var result []string
	for _, v := range values {
		result = append(result, Sanitize(pattern, v))
	}
	return result
}
