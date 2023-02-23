package pgpkg

// This file runs the tests. Tests are a bit more complicated than you might expect because
// we don't want any results to be written to the database, but we need to maintain
// various states so we can report errors and get stack traces.
//
// Nothing is overly complex; but it's not as simple as just running the units.

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Tests struct {
	Bundle
	state      *stmtApplyState
	NamedTests map[string]*Statement
}

func (p *Package) loadTests(path string) (*Tests, error) {
	bundle, err := p.loadBundle(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Tests{}, nil
		}

		return nil, err
	}

	tests := &Tests{
		Bundle: *bundle,
	}

	return tests, nil
}

func (t *Tests) parse() error {
	var pending []*Statement

	namedTests := make(map[string]*Statement)
	definitions := make(map[string]*Statement)

	for _, u := range t.Units {
		if t.Package.Options.Verbose {
			fmt.Println("parsing tests", u.Location())
		}

		if err := u.Parse(); err != nil {
			return fmt.Errorf("unable to parse tests: %w", err)
		}

		for _, stmt := range u.Statements {
			obj, err := stmt.GetObject()
			if err != nil {
				return err
			}

			if obj.ObjectType != "function" {
				return PKGErrorf(stmt, nil, "only functions can be defined in tests; %s %s", obj.ObjectType, obj.ObjectName)
			}

			if len(obj.ObjectArgs) != 0 {
				return PKGErrorf(stmt, nil, "test functions cannot receive arguments: %s %s", obj.ObjectType, obj.ObjectName)
			}

			// Check for duplicate test definitions. This can be a subtle bug because
			// all the statements are probably "create or replace".
			objName := obj.ObjectType + ":" + obj.ObjectName
			dupeStmt, dupe := definitions[objName]
			if dupe {
				return PKGErrorf(stmt, nil,
					"duplicate declaration for %s %s; also defined in %s",
					obj.ObjectType, obj.ObjectName, dupeStmt.Location())
			}
			definitions[objName] = stmt

			// Get the underlying name of the function
			fname := strings.ToLower(strings.TrimPrefix(obj.ObjectName, obj.ObjectSchema+"."))

			if strings.HasPrefix(fname, "test_") {
				t.Package.StatTestCount++
				namedTests[obj.ObjectName] = stmt
			}

			pending = append(pending, stmt)
		}
	}

	t.NamedTests = namedTests
	t.state = &stmtApplyState{pending: pending}
	return nil
}

// testStmt is the statement containing the test function.
// testStmt was executed when the tests were parsed, so it is only used to work out where
// problems might have happened.
func (t *Tests) runTest(tx *sql.Tx, testName string, testStmt *Statement) error {
	if _, spErr := tx.Exec("savepoint unittest"); spErr != nil {
		return fmt.Errorf("unable to begin savepoint for test %s: %w", testName, spErr)
	}

	cmd := fmt.Sprintf("select %s", testName)
	_, testErr := tx.Exec(cmd)

	_, rberr := tx.Exec("rollback to savepoint unittest")
	if rberr != nil {
		panic(rberr)
	}

	if testErr == nil {
		if t.Package.Options.Verbose {
			fmt.Println("[PASS]", testName)
		}
		return nil
	}

	pe := PKGErrorf(testStmt, testErr, "test failed: %s", testName)
	pe.Context = getErrorContext(tx, cmd, testErr)
	tx = nil
	return pe

	return testErr
}

func (t *Tests) Run(tx *sql.Tx) error {

	// Rollback, and return either the error or an error from the rollback.
	defer func() {
		if tx != nil {
			_, rberr := tx.Exec("rollback to savepoint test")
			if rberr != nil {
				panic(rberr)
			}
		}
	}()

	// Create a savepoint for the entire set of tests.
	_, err := tx.Exec("savepoint test")
	if err != nil {
		return fmt.Errorf("unable to begin test savepoint: %w", err)
	}

	// Parse all the functions.
	err = t.parse()
	if err != nil {
		return err
	}

	err = applyState(tx, t.state)
	if err != nil {
		return err
	}

	for testName, testStmt := range t.NamedTests {
		err := t.runTest(tx, testName, testStmt)

		if err != nil {
			return err
		}
	}

	return nil
}
