package pgpkg

import (
	"database/sql"
	"fmt"
	"strings"
)

type PkgTx struct {
	*sql.Tx
	Verbose bool
}

func (t *PkgTx) Exec(query string, args ...any) (sql.Result, error) {
	t.logQuery(query, args)
	return t.Tx.Exec(query, args...)
}

func (t *PkgTx) Query(query string, args ...any) (*sql.Rows, error) {
	t.logQuery(query, args)
	return t.Tx.Query(query, args...)
}

func (t *PkgTx) QueryRow(query string, args ...any) *sql.Row {
	t.logQuery(query, args)
	return t.Tx.QueryRow(query, args...)
}

func (t *PkgTx) logQuery(query string, args []any) {
	if !t.Verbose {
		return
	}

	logText := query
	newLine := strings.IndexRune(logText, '\n')
	if newLine >= 0 {
		logText = logText[:newLine]
	}

	if args == nil || len(args) == 0 {
		fmt.Println(logText)
	} else {
		fmt.Println(query, args)
	}
}

