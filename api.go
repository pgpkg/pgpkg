package pgpkg

// An API is a kind of bundle that manages objects that implement domain logic,
// which can change over time as the schema grows and changes. It is, in effect,
// the API for manipulating the schema.
//
// APIs consist only of stored functions, views, and triggers. We might add additional
// objects over time. APIs do not include tables, indexes or other similar objects.
//
// API bundles don't care about build units; they can be considered instead to be a random collection
// of CREATE statements. The order in which the CREATE statements is executed is initially set by the
// order in which they are encountered (ie, lexically within build units), but pgpkg will re-order the
// statements until a build succeeds or until it fails because progress can't be made.

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
)

type API struct {
	*Bundle
	state *stmtApplyState
}

func (p *Package) loadAPI(path string) (*API, error) {
	bundle, err := p.loadBundle(APIBundleDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	app := &API{
		Bundle: bundle,
	}

	return app, nil
}

// Track the statements as we attempt to find an ordering that works.
type stmtApplyState struct {
	pending []*Statement
	failed  []*Statement
	success []*Statement
}

func (a *API) Parse() error {
	var pending []*Statement
	definitions := make(map[string]*Statement)

	for _, u := range a.Units {
		if a.Package.Options.Verbose {
			fmt.Println("parsing api", u.Path)
		}
		if err := u.Parse(); err != nil {
			return fmt.Errorf("unable to parse API: %w", err)
		}

		for _, stmt := range u.Statements {
			obj, err := stmt.GetObject()
			if err != nil {
				return err
			}

			if obj.ObjectType == "function" {
				// Rewrite the statement to set the schema and security options.
				err = rewrite(stmt)
				if err != nil {
					return err
				}
			}

			// Check for duplicate definitions in the API. This can be a subtle bug because
			// all the statements are probably "create or replace".
			objName := obj.ObjectType + ":" + obj.ObjectName
			dupeStmt, dupe := definitions[objName]
			if dupe {
				return PKGErrorf(stmt, nil,
					"duplicate declaration for %s %s; also defined in %s",
					obj.ObjectType, obj.ObjectName, dupeStmt.Location())
			}
			definitions[objName] = stmt

			pkg := a.Package
			switch obj.ObjectType {
			case "function":
				pkg.StatFuncCount++
			case "view":
				pkg.StatViewCount++
			case "trigger":
				pkg.StatTriggerCount++
			}

			pending = append(pending, stmt)
		}
	}

	a.state = &stmtApplyState{pending: pending}
	return nil
}

// ExecAll attempts to run each of the statements in the pending list.
// Each statement is run in a savepoint. Any statements that did not execute
// successfully are returned.
func execAll(tx *sql.Tx, state *stmtApplyState) error {
	for _, stmt := range state.pending {
		ok, err := stmt.Try(tx)
		if !ok {
			return err
		}

		if err != nil {
			// this is normal, and will happen if there is a missing dependency. We will
			// try the statement again in the next pass.
			state.failed = append(state.failed, stmt)
			continue
		}

		// It worked; keep track of the order
		state.success = append(state.success, stmt)
		//fmt.Println("OK:", stmt.Headline())
	}

	return nil
}

type stmtStoredState struct {
	objType string
	objName string
}

// loadState returns the state objects in reverse order from how they were created.
// this should make dumping objects faster.
func (a *API) loadState(tx *sql.Tx) ([]stmtStoredState, error) {
	rows, err := tx.Query("select obj_type, obj_name from pgpkg.api where pkg=$1 order by seq desc",
		a.Package.Name)
	if err != nil {
		return nil, PKGErrorf(a, err, "unable to load API state")
	}

	var stateList []stmtStoredState

	for rows.Next() {
		state := stmtStoredState{}
		if err := rows.Scan(&state.objType, &state.objName); err != nil {
			return nil, PKGErrorf(a, err, "error during load of API state")
		}
		stateList = append(stateList, state)

	}

	return stateList, nil
}

func applyState(tx *sql.Tx, state *stmtApplyState) error {
	for {
		lenPending := len(state.pending)
		if lenPending == 0 {
			break
		}

		err := execAll(tx, state)
		if err != nil {
			return err
		}

		// Replace the pending list with the failed list, and maybe try again.
		state.pending = state.failed
		state.failed = nil

		if len(state.pending) == lenPending {
			ps := state.pending[0]
			return PKGErrorf(ps, ps.Error, "unable to install API")
		}
	}

	return nil
}

// Purge (drop) all the managed API objects. This is performed
// recursively to ensure that dependent objects are also deleted, if possible.
// We don't use CASCADE with drops to ensure that any other scheme that inadvertently relies
// on API functions is not damaged by the purge.
func (a *API) purge(tx *sql.Tx) error {
	var pending []*Statement

	state, err := a.loadState(tx)
	if err != nil {
		return err
	}

	for _, obj := range state {
		pending = append(pending, &Statement{
			Source:     fmt.Sprintf("drop %s if exists %s", obj.objType, obj.objName),
			LineNumber: 1,
		})
	}

	purgeState := &stmtApplyState{
		pending: pending,
	}

	return applyState(tx, purgeState)
}

// Update the database with the new state of the API.
func (a *API) updateState(tx *sql.Tx) error {
	_, err := tx.Exec("delete from pgpkg.api where pkg=$1", a.Bundle.Package.Name)
	if err != nil {
		return fmt.Errorf("unable to remove existing state: %w", err)
	}

	for seq, stmt := range a.state.success {
		obj, err := stmt.GetObject()
		if err != nil {
			return err
		}

		if obj != nil {
			_, err = tx.Exec(
				"insert into pgpkg.api (pkg, seq, obj_type, obj_name) "+
					"values ($1, $2, $3, $4)", a.Bundle.Package.Name, seq, obj.ObjectType, obj.ObjectName)
			if err != nil {
				return fmt.Errorf("unable to update package state: %w", err)
			}
		}
	}

	return nil
}

func (a *API) Location() string {
	return a.Package.Name
}

func (a *API) DefaultContext() *PKGErrorContext {
	return nil
}

// Apply performs the SQL required to create the objects listed in the
// API object, to register them in the pgpkg.object table.
// Since objects in an API may depend on one another, this
// function starts with a list of the statements to be executed,
// and attempts to execute them one at a time.
//
// Each statement is executed in a savepoint. If a statement fails,
// we skip over it and keep trying.
//
// The apply function will keep running until it's unable to create
// any statement, after which it will terminate.
//
// TODO: use the API table to get hints about the order of
// statement execution, which might speed things up.
//
// Returns the statements in the order they were successfully executed.
func (a *API) Apply(tx *sql.Tx) error {
	if a.state == nil {
		panic("please call API.Parse() before calling API.Apply()")
	}

	return applyState(tx, a.state)
}