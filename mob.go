package pgpkg

import (
	"fmt"
)

// MOB (managed object bundle) is a kind of bundle that manages objects that implement domain logic,
// which can change over time as the schema grows and changes.
//
// MOBs consist only of stored functions, views, and triggers. We might add additional
// objects over time. MOBs will never include tables, indexes or other similar objects.
//
// MOBs only care about the contents of build units, but not the units themselves; MOBs can be
// considered instead to be a random collection of CREATE statements. The order in which the CREATE
// statements is executed is initially set by the order in which they are encountered (ie, lexically
// within build units), but pgpkg will re-order the statements until a build succeeds or until it
// fails because progress can't be made.

type MOB struct {
	*Bundle
	state *stmtApplyState
}

// Track the statements as we attempt to find an ordering that works.
type stmtApplyState struct {
	pending []*Statement
	failed  []*Statement
	success []*Statement
}

func (m *MOB) Parse() error {
	var pending []*Statement
	definitions := make(map[string]*Statement)

	for _, u := range m.Units {
		if Options.Verbose {
			Verbose.Println("parsing MOB", u.Location())
		}
		if err := u.Parse(); err != nil {
			return fmt.Errorf("unable to parse MOB: %w", err)
		}

		for _, stmt := range u.Statements {
			obj, err := stmt.GetManagedObject()
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

			// Check for duplicate definitions in the MOB. This can be a subtle bug because
			// all the statements are probably "create or replace".
			objName := obj.ObjectType + ":" + obj.ObjectName
			dupeStmt, dupe := definitions[objName]
			if dupe {
				return PKGErrorf(stmt, nil,
					"duplicate declaration for %s %s; also defined in %s",
					obj.ObjectType, obj.ObjectName, dupeStmt.Location())
			}
			definitions[objName] = stmt

			pkg := m.Package
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

	m.state = &stmtApplyState{pending: pending}
	return nil
}

// ExecAll attempts to run each of the statements in the pending list.
// Each statement is run in a savepoint.
func execAll(tx *PkgTx, state *stmtApplyState) error {
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
	}

	return nil
}

type stmtStoredState struct {
	objType string
	objName string
}

func (s *stmtStoredState) getDropStatement() string {
	switch s.objType {
	case "function", "view", "trigger":
		return fmt.Sprintf("drop %s if exists %s", s.objType, s.objName)
	case "comment on function", "comment on view", "comment on column":
		return fmt.Sprintf("%s %s is null", s.objType, s.objName)
	case "cast":
		return fmt.Sprintf("drop cast (%s)", s.objName)
	case "unknown":
		return ""
	}

	panic(fmt.Errorf("unknown object type: %s", s.objType))
}

// loadState returns the state objects in reverse order from how they were created.
// this should make dumping objects faster.
func (m *MOB) loadState(tx *PkgTx) ([]*stmtStoredState, error) {
	rows, err := tx.Query("select obj_type, obj_name from pgpkg.managed_object where pkg=$1 order by seq desc",
		m.Package.Name)
	if err != nil {
		return nil, PKGErrorf(m, err, "unable to load MOB state")
	}

	var stateList []*stmtStoredState

	for rows.Next() {
		state := &stmtStoredState{}
		if err := rows.Scan(&state.objType, &state.objName); err != nil {
			return nil, PKGErrorf(m, err, "error during load of MOB state")
		}
		stateList = append(stateList, state)

	}

	return stateList, nil
}

func applyState(tx *PkgTx, state *stmtApplyState) error {
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

		// If we weren't able to make any progress at all, then something's wrong.
		if len(state.pending) == lenPending {
			allErrors := []*PKGError{}
			for _, pending := range state.pending {
				if pending.Error != nil {
					allErrors = append(allErrors, PKGErrorf(pending, pending.Error, "unable to install MOB"))
				}
			}

			if len(allErrors) > 1 {
				allErrors[0].Errors = allErrors[1:]
			}

			return allErrors[0]
		}
	}

	return nil
}

// Purge (drop) all the managed MOB objects. This is performed
// recursively to ensure that dependent objects are also deleted, if possible.
// We don't use CASCADE with drops to ensure that any other scheme that inadvertently relies
// on MOB functions is not damaged by the purge.
func (m *MOB) purge(tx *PkgTx) error {
	var pending []*Statement

	state, err := m.loadState(tx)
	if err != nil {
		return err
	}

	for _, obj := range state {
		pending = append(pending, &Statement{
			Source:     obj.getDropStatement(), //fmt.Sprintf("drop %s if exists %s", obj.objType, obj.objName),
			LineNumber: 1,
		})
	}

	purgeState := &stmtApplyState{
		pending: pending,
	}

	return applyState(tx, purgeState)
}

// Update the pgpkg.managed_object table with the new state of the MOB, by deleting existing
// entries and inserting new ones from the list of successful MOBs processed.
func (m *MOB) updateState(tx *PkgTx) error {
	_, err := tx.Exec("delete from pgpkg.managed_object where pkg=$1", m.Bundle.Package.Name)
	if err != nil {
		return fmt.Errorf("unable to remove existing state: %w", err)
	}

	for seq, stmt := range m.state.success {
		obj, err := stmt.GetManagedObject()
		if err != nil {
			return err
		}

		if obj != nil {
			_, err = tx.Exec(
				"insert into pgpkg.managed_object (pkg, seq, obj_type, obj_name) "+
					"values ($1, $2, $3, $4)", m.Bundle.Package.Name, seq, obj.ObjectType, obj.ObjectName)
			if err != nil {
				return fmt.Errorf("unable to update package state: %w", err)
			}
		}
	}

	return nil
}

func (m *MOB) Location() string {
	return m.Package.Name
}

func (m *MOB) DefaultContext() *PKGErrorContext {
	return nil
}

// Apply performs the SQL required to create the objects listed in the
// MOB object, to register them in the pgpkg.object table.
// Since objects in an MOB may depend on one another, this
// function starts with a list of the statements to be executed,
// and attempts to execute them one at a time.
//
// Each statement is executed in a savepoint. If a statement fails,
// we skip over it and keep trying.
//
// The apply function will keep running until it's unable to create
// any statement, after which it will terminate.
func (m *MOB) Apply(tx *PkgTx) error {
	if m.state == nil {
		panic("please call MOB.Parse() before calling MOB.Apply()")
	}

	return applyState(tx, m.state)
}

func (m *MOB) PrintInfo(w InfoWriter) {
	w.Println("Managed Object Bundle")
	m.Bundle.PrintInfo(w)
}
