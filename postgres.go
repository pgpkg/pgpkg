package pgpkg

import (
	"database/sql"
	"fmt"
	"github.com/gookit/color"
	"github.com/lib/pq"
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

func Open(conninfo string, options *Options) (*sql.DB, error) {
	base, err := pq.NewConnector(conninfo)
	if err != nil {
		return nil, fmt.Errorf("connection to database: %w", err)
	}

	// Wrap the connector to simply print out the message. Capture the options
	// so we can enable verbose, etc.
	connector := pq.ConnectorWithNoticeHandler(base,
		func(err *pq.Error) {
			noticeHandler(options, err)
		})

	// Open the database
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	return db, nil
}

func LogQuieter() {
	logVolume++
}

func LogLouder() {
	logVolume--
}
