package pgpkg

import (
	"fmt"
	"github.com/lib/pq"
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"regexp"
	"strconv"
	"strings"
)

// Statement is a parsed SQL statement within a unit.
type Statement struct {
	Unit       *Unit             // Unit this statement appears in
	LineNumber int               // Line number within the Unit
	Source     string            // The actual SQL
	Tree       *pg_query.RawStmt // Parsed SQL statement.
	Error      error             // The most recent result from processing the statement.

	object *ManagedObject // Cached result of GetManagedObject()
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

// Try executes a statement in a savepoint. This allows us to find context
// if statement execution fails.
//
// Returns true if the statement succeeded, or true-with-error if it failed
// but could be retried (this depends on where the error occurred). Returns
// false if an error occurred that was not related to statement execution.
//
// If an error occurs while executing the statement, the statement's Error field is also set.
func (s *Statement) Try(tx *PkgTx) (bool, error) {
	_, err := tx.Exec("savepoint statement")
	if err != nil {
		return false, fmt.Errorf("unable to begin savepoint: %w", err)
	}

	_, err = tx.Exec(s.Source)
	if err != nil {
		_, rberr := tx.Exec("rollback to savepoint statement")
		if rberr != nil {
			return false, fmt.Errorf("unable to rollback to savepoint: %w", rberr)
		}
		pkgError := PKGErrorf(s, err, "unable to execute statement")
		s.Error = pkgError

		// Attempt to find some additional context for this error.
		pkgError.Context = s.getErrorContext(tx, err)

		return true, pkgError
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

var linePattern = regexp.MustCompile("line ([0-9]+)")

// Work out the source and line number of a runtime error, either by looking at the statement
// source or looking in the database for a function definition. Returns nil if the context
// can't be worked out.
func getRuntimeContext(tx *PkgTx, source string, location string) *PKGErrorContext {
	lines := linePattern.FindStringSubmatch(location)
	if lines == nil || len(lines) != 2 {
		return nil
	}

	// Being unable to parse the line number returned by an error isn't itself an error,
	// it's just inconvenient.
	lineNumber, err := strconv.Atoi(lines[1])
	if err != nil {
		return nil
	}

	// If the error was from inline code, then the context comes from the statement.
	if strings.Contains(location, "inline_code_block") {
		return &PKGErrorContext{
			Source:     source,
			Location:   location,
			LineNumber: lineNumber,
		}
	}

	// If the error identifies a specific function, we can look it up in the database
	// and use that as the context.
	if strings.Contains(location, "function") {
		functionSource, err := getFunctionSource(tx, location)
		if err != nil {
			return &PKGErrorContext{
				Location:   location,
				LineNumber: lineNumber,
			}
		}

		return &PKGErrorContext{
			Source:     functionSource,
			Location:   location,
			LineNumber: lineNumber,
		}
	}

	// Otherwise we don't seem to be able to find the context. C'est la vie.
	return nil
}

func getErrorContext(tx *PkgTx, source string, err error) *PKGErrorContext {
	var where string

	// If it's not a pq.Error, then the context comes from the statement itself.
	pgerr, ok := err.(*pq.Error)
	if !ok {
		return nil
	}

	where = pgerr.Where

	// If "where" isn't set, we can use the position to determine the line number.
	if len(where) == 0 {
		position, _ := strconv.Atoi(pgerr.Position)
		return &PKGErrorContext{
			Source:     source,
			LineNumber: 1 + strings.Count(source[:position], "\n"),
		}
	}

	// If "where" is set, it will contain the whole stack.
	// In this case, return the contexts.
	locations := strings.Split(where, "\n")
	var lastContext *PKGErrorContext
	for index := len(locations) - 1; index >= 0; index-- {
		ec := getRuntimeContext(tx, source, locations[index])
		if ec != nil {
			if lastContext != nil {
				ec.Next = lastContext
			}
			lastContext = ec
		}
	}

	if lastContext != nil {
		return lastContext
	}

	// Couldn't find anything, so don't return a context.
	return nil
}

func (s *Statement) getErrorContext(tx *PkgTx, err error) *PKGErrorContext {
	ec := getErrorContext(tx, s.Source, err)
	if ec != nil {
		return ec
	}

	return &PKGErrorContext{
		Source:     s.Unit.Source,
		LineNumber: s.LineNumber,
	}
}

func (s *Statement) Location() string {
	return fmt.Sprintf("%s:%d", s.Unit.Location(), s.LineNumber)
}

func (s *Statement) DefaultContext() *PKGErrorContext {
	return &PKGErrorContext{
		Source:     s.Source,
		LineNumber: s.LineNumber,
	}
}
