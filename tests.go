package pgpkg

// This file runs the tests. Tests are a bit more complicated than you might expect because
// we don't want any results to be written to the database, but we need to maintain
// various states so we can report errors and get stack traces.
//
// Nothing is overly complex; but it's not as simple as just running the units.

import (
	"fmt"
	"os"
	"strings"
)

type Tests struct {
	*Bundle
	state      *stmtApplyState
	NamedTests map[string]*Statement
}

func (t *Tests) parse() error {
	var pending []*Statement

	namedTests := make(map[string]*Statement)
	definitions := make(map[string]*Statement)

	for _, u := range t.Units {
		if Options.Verbose {
			Verbose.Println("parsing tests", u.Location())
		}

		if err := u.Parse(); err != nil {
			return fmt.Errorf("unable to parse tests: %w", err)
		}

		for _, stmt := range u.Statements {
			obj, err := stmt.GetManagedObject()
			if err != nil {
				return err
			}

			if obj.ObjectType != "function" {
				return PKGErrorf(stmt, nil, "only functions can be defined in tests; %s %s", obj.ObjectType, obj.ObjectName)
			}

			// Rewrite the statement to set the schema and security options.
			err = rewrite(stmt)
			if err != nil {
				return err
			}

			// Get the unqualified name of the function.
			fname := strings.ToLower(strings.TrimPrefix(obj.ObjectName, obj.ObjectSchema+"."))
			argIndex := strings.IndexRune(fname, '(')
			fname = fname[:argIndex]
			//fname = strings.TrimSuffix(fname, "()") // FIXME: BUG: test funcs with args won't end in ()!!!!

			// test_function_name is deprecated. Use function_name_test, to match filename.
			isTestFunction := strings.HasSuffix(fname, "_test")
			if !isTestFunction {
				isTestFunction = strings.HasPrefix(fname, "test_")
				if isTestFunction {
					fmt.Fprintf(os.Stderr, "warning: test name %s is deprecated; use %s_test instead\n", fname, fname[5:])
				}
			}

			if isTestFunction && len(obj.ObjectArgs) != 0 {
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

			if isTestFunction {
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
func (t *Tests) runTest(tx *PkgTx, testName string, testStmt *Statement) error {
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
		if Options.ShowTests {
			Stdout.Println("  [pass]", testName)
		}
		return nil
	}

	if Options.ShowTests {
		Stdout.Println("* [FAIL]", testName)
	}

	pe := PKGErrorf(testStmt, testErr, "test failed: %s", testName)
	pe.Context = getErrorContext(tx, cmd, testErr)
	tx = nil
	return pe

	return testErr
}

func (t *Tests) Run(tx *PkgTx) error {

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
		if Options.IncludePattern != nil {
			if !Options.IncludePattern.MatchString(testName) {
				if Options.ShowSkipped {
					Stdout.Println("- [skip]", testName)
				}
				continue
			}
		}

		if Options.ExcludePattern != nil {
			if Options.ExcludePattern.MatchString(testName) {
				if Options.ShowSkipped {
					Stdout.Println("- [skip]", testName)
				}
				continue
			}
		}

		err := t.runTest(tx, testName, testStmt)

		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Tests) PrintInfo(w InfoWriter) {
	w.Println("Test Bundle")
	t.Bundle.PrintInfo(w)
}
