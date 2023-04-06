package pgpkg

import (
	"fmt"
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"strings"
)

// ManagedObject refers to a managed database object, with a schema name,
// object name and object type.
type ManagedObject struct {
	ObjectSchema string
	ObjectType   string
	ObjectName   string
	ObjectArgs   []string
}

func getParamType(fp *pg_query.FunctionParameter) string {
	argType := fp.ArgType
	typeName := QualifiedName(fp.ArgType.Names)
	if argType.ArrayBounds != nil {
		for _ = range argType.ArrayBounds {
			typeName = typeName + "[]"
		}
	}

	return typeName
}

func (s *Statement) getFunctionObject() (*ManagedObject, error) {
	createFunctionStmt := s.Tree.Stmt.GetCreateFunctionStmt()
	pkg := s.Unit.Bundle.Package

	var args []string
	for _, arg := range createFunctionStmt.Parameters {
		fp := arg.GetFunctionParameter()
		if fp.Mode == pg_query.FunctionParameterMode_FUNC_PARAM_IN ||
			fp.Mode == pg_query.FunctionParameterMode_FUNC_PARAM_INOUT ||
			fp.Mode == pg_query.FunctionParameterMode_FUNC_PARAM_DEFAULT {
			args = append(args, fmt.Sprintf("%s %s", fp.Name, getParamType(fp)))
		}
	}
	schema := AsString(createFunctionStmt.Funcname[0])
	if schema == "" {
		return nil, PKGErrorf(s, nil, "no function schema declared")
	}

	if !pkg.isValidSchema(schema) {
		return nil, PKGErrorf(s, nil, "function schema %s is not declared in package", schema)
	}

	return &ManagedObject{
		ObjectSchema: schema,
		ObjectType:   "function",
		ObjectName:   fmt.Sprintf("%s.%s(%s)", schema, AsString(createFunctionStmt.Funcname[1]), strings.Join(args, ",")),
		ObjectArgs:   args,
	}, nil
}

func (s *Statement) getTriggerObject() (*ManagedObject, error) {
	createTrigStmt := s.Tree.Stmt.GetCreateTrigStmt()
	pkg := s.Unit.Bundle.Package

	name := createTrigStmt.Trigname
	schema := createTrigStmt.Relation.Schemaname
	table := createTrigStmt.Relation.Relname

	if schema == "" {
		return nil, PKGErrorf(s, nil, "no schema declared on trigger table")
	}

	if !pkg.isValidSchema(schema) {
		return nil, PKGErrorf(s, nil, "trigger table schema %s is not declared in package", schema)
	}

	return &ManagedObject{
		ObjectSchema: schema,
		ObjectType:   "trigger",
		ObjectName:   fmt.Sprintf("%s on %s.%s", name, schema, table),
	}, nil
}

func (s *Statement) getViewObject() (*ManagedObject, error) {
	viewStmt := s.Tree.Stmt.GetViewStmt()
	pkg := s.Unit.Bundle.Package

	schema := viewStmt.View.Schemaname
	name := viewStmt.View.Relname

	if schema == "" {
		return nil, PKGErrorf(s, nil, "no schema declared on view")
	}

	if !pkg.isValidSchema(schema) {
		return nil, PKGErrorf(s, nil, "view schema %s is not declared in package", schema)
	}

	return &ManagedObject{
		ObjectSchema: schema,
		ObjectType:   "view",
		ObjectName:   fmt.Sprintf("%s.%s", schema, name)}, nil
}

// GetManagedObject returns identifying information about an object from a CREATE
// statement, such as function, view or trigger. NOTE: This functon
// might not support all object types, but you can add more as needed.
//
// The result is cached since it's used repeatedly during MOB processing.
func (s *Statement) GetManagedObject() (*ManagedObject, error) {

	if s.object != nil {
		return s.object, nil
	}

	stmt := s.Tree.Stmt
	var err error

	switch {
	case stmt.GetCreateFunctionStmt() != nil:
		s.object, err = s.getFunctionObject()

	case stmt.GetCreateTrigStmt() != nil:
		s.object, err = s.getTriggerObject()

	case stmt.GetViewStmt() != nil:
		s.object, err = s.getViewObject()
	}

	if err != nil {
		return nil, err
	}

	if s.object != nil {
		return s.object, nil
	}

	return nil, PKGErrorf(s, nil, "only functions, triggers and views are supported in MOB")
}
