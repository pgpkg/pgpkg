package pgpkg

import (
	"database/sql"
	"fmt"
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"strings"
)

// Statement is a parsed SQL statement within a unit.
type Statement struct {
	Unit       *Unit             // Unit this statement appears in
	LineNumber int               // Line number within the Unit
	Source     string            // The actual SQL
	Tree       *pg_query.RawStmt // Parsed SQL statement.
	Error      error             // The most recent result from processing the statement.

	object *Object	// Cached result of GetObject()
}

// AsString is a utility function to get the string value of a node.
func AsString(node *pg_query.Node) string {
	return node.GetString_().GetSval()
}

func QualifiedName(nodes []*pg_query.Node) string {
	var names []string
	for _, node := range nodes {
		names = append(names, AsString(node))
	}
	return strings.Join(names, ".")
}

func (s *Statement) Exec(tx *sql.Tx) error {
	_, err := tx.Exec(s.Source)

	// FIXME: include location information
	return err
}

// Try attempts to execute the statement in a savepoint.
// Returns true if the statement succeeded, or true-with-error if it failed
// but could be retried. Returns false if an error occurred that was not related to
// statement execution.
//
// If an error occurs while executing the statement, the statement's Error flag is also set.
func (s *Statement) Try(tx *sql.Tx) (bool, error) {

	_, err := tx.Exec("savepoint statement")
	if err != nil {
		return false, err
	}

	s.Error = s.Exec(tx)
	if s.Error != nil {
		_, rberr := tx.Exec("rollback to savepoint statement")
		if rberr != nil {
			return false, fmt.Errorf("unable to rollback to savepoint: %w", rberr)
		}
		return true, s.Error
	}

	_, relerr := tx.Exec("release savepoint statement")
	if relerr != nil {
		return false, fmt.Errorf("unable to release savepoint: %w", relerr)
	}

	return true, nil
}

// Headline returns the first line of the statement, eg, to provide context
// during debugging and logging.
func (s *Statement) Headline() string {
	if s.Unit != nil {
		lines := strings.Split(s.Unit.Source, "\n")
		return lines[s.LineNumber-1]
	} else {
		lines := strings.Split(s.Source, "\n")
		return lines[s.LineNumber-1]
	}
}

func (s *Statement) Location() string {
	return fmt.Sprintf("%s:%d", s.Unit.Location(), s.LineNumber)
}
