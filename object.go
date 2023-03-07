package pgpkg

import (
	"fmt"
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"strings"
)

// Object refers to a managed database object, with a schema name,
// object name and obejct type.
type Object struct {
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

func (s *Statement) getFunctionObject() (*Object, error) {
	createFunctionStmt := s.Tree.Stmt.GetCreateFunctionStmt()
	pkgSchema := s.Unit.Bundle.Package.SchemaName

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

	if schema != pkgSchema {
		return nil, PKGErrorf(s, nil, "declared schema %s does not match package schema %s", schema, pkgSchema)
	}

	return &Object{
		ObjectSchema: schema,
		ObjectType:   "function",
		ObjectName:   fmt.Sprintf("%s.%s(%s)", schema, AsString(createFunctionStmt.Funcname[1]), strings.Join(args, ",")),
		ObjectArgs:   args,
	}, nil
}

func (s *Statement) getTriggerObject() (*Object, error) {
	createTrigStmt := s.Tree.Stmt.GetCreateTrigStmt()
	pkgSchema := s.Unit.Bundle.Package.SchemaName

	name := createTrigStmt.Trigname
	schema := createTrigStmt.Relation.Schemaname
	table := createTrigStmt.Relation.Relname

	if schema == "" {
		return nil, PKGErrorf(s, nil, "no schema declared on trigger table")
	}

	if schema != pkgSchema {
		return nil, PKGErrorf(s, nil, "trigger table schema %s does not match package schema %s", schema, pkgSchema)
	}

	//name := GetTriggerDeclaration(createTrigStmt)
	return &Object{
		ObjectSchema: schema,
		ObjectType:   "trigger",
		ObjectName:   fmt.Sprintf("%s on %s.%s", name, schema, table),
	}, nil
}

func (s *Statement) getViewObject() (*Object, error) {
	viewStmt := s.Tree.Stmt.GetViewStmt()
	pkgSchema := s.Unit.Bundle.Package.SchemaName

	schema := viewStmt.View.Schemaname
	name := viewStmt.View.Relname

	if schema == "" {
		return nil, PKGErrorf(s, nil, "no schema declared on view")
	}

	if schema != pkgSchema {
		return nil, PKGErrorf(s, nil, "view schema %s does not match package schema %s", schema, pkgSchema)
	}

	return &Object{
		ObjectSchema: schema,
		ObjectType:   "view",
		ObjectName:   fmt.Sprintf("%s.%s", schema, name)}, nil
}

// GetObject returns identifying information about an object from a CREATE
// statement, such as function, view or trigger. NOTE: This functon
// might not support all object types, but you can add more as needed.
//
// The result is cached since it's used repeatedly during MOB processing.
func (s *Statement) GetObject() (*Object, error) {

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
