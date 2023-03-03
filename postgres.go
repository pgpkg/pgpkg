package pgpkg

import (
	"database/sql"
	"fmt"
	"github.com/gookit/color"
	"github.com/lib/pq"
	"regexp"
	"strings"
)

// The "quiet level" is incremented or decremented to supress messages from the database.
// This help to stop database generated messages that will make users think something is wrong
// when it isn't. These messages are great during debugging but should be silenced during
// operations that will generate them spuriously; sometimes PG itself emits them when we
// don't want them.
var logVolume = 0

func noticeHandler(options *Options, err *pq.Error) {
	// Don't allow warnings to be quiet.
	if err.Severity == "WARNING" {
		color.Red.Printf("[%s]: %s\n", err.Severity, err.Message)
	} else {
		if logVolume == 0 || options.Verbose {
			color.Bold.Printf("[%s]: %s\n", strings.ToLower(err.Severity), err.Message)
		}
	}
}

func LogQuieter() {
	logVolume--
}

func LogLouder() {
	logVolume++
}

var fnamePattern = regexp.MustCompile("function ([a-z_][a-z0-9_.]*[(].*[)])")

// Get the source of a function from the database itself, based on the "where" field
// of a pgsql error.
func getFunctionSource(tx *sql.Tx, where string) (string, error) {

	// The where string should contain a function name.
	fnames := fnamePattern.FindStringSubmatch(where)
	if len(fnames) != 2 {
		return "", fmt.Errorf("can't identify function in error detail")
	}

	// Convert the function name into an OID
	foidRow := tx.QueryRow("select $1::pg_catalog.regprocedure::pg_catalog.oid", fnames[1])
	var foid int
	err := foidRow.Scan(&foid)
	if err != nil {
		return "", fmt.Errorf("error looking up function name: %w", err)
	}

	// Look up the OID to get the source of the function.
	var src string
	srcRow := tx.QueryRow("select prosrc from pg_catalog.pg_proc where OID=$1", foid)
	err = srcRow.Scan(&src)
	if err != nil {
		return "", fmt.Errorf("error looking up function source: %w", err)
	}

	return src, nil
}
