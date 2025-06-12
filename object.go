package pgpkg

import (
	"fmt"
	pg_query "github.com/pganalyze/pg_query_go/v6"
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

func getTypeName(name *pg_query.TypeName) string {
	typeName := QualifiedName(name.Names)
	if name.ArrayBounds != nil {
		for range name.ArrayBounds {
			typeName = typeName + "[]"
		}
	}

	return typeName
}

func getParamType(fp *pg_query.FunctionParameter) string {
	return getTypeName(fp.ArgType)
}

// Convert the identifier to a quoted string. If the string contains ", these are replaced with "".
func quote(v string) string {
	return "\"" + strings.Replace(v, "\"", "\"\"", -1) + "\""
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

		// note that pg_analyze_go doesn't seem to support quoted argument names in functions. To avoid assumptions
		// about future pg_analyze_go future behaviour, we won't support them here either.
		ObjectName: fmt.Sprintf("%s.%s(%s)", quote(schema), quote(AsString(createFunctionStmt.Funcname[1])), strings.Join(args, ",")),
		ObjectArgs: args,
	}, nil
}

func (s *Statement) getCastObject() (*ManagedObject, error) {
	createCastStmt := s.Tree.Stmt.GetCreateCastStmt()
	return &ManagedObject{
		ObjectSchema: "public",
		ObjectType:   "cast",
		ObjectName:   getTypeName(createCastStmt.Sourcetype) + " as " + getTypeName(createCastStmt.Targettype),
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
		ObjectName:   fmt.Sprintf("%s on %s.%s", quote(name), quote(schema), quote(table)),
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
		ObjectName:   fmt.Sprintf("%s.%s", quote(schema), quote(name))}, nil
}

func (s *Statement) getCommentObject() (*ManagedObject, error) {
	commentStmt := s.Tree.Stmt.GetCommentStmt()
	ot := commentStmt.Objtype

	switch ot {
	case pg_query.ObjectType_OBJECT_FUNCTION:
		// WARNING: this does not capture the entire function name, which should include parameters.
		//  Capturing the whole name with parameters will probably require a minor refactor.
		funcObject := commentStmt.Object.GetObjectWithArgs()
		return &ManagedObject{
			ObjectSchema: AsString(funcObject.Objname[0]),
			ObjectType:   "comment on function",
			ObjectName:   fmt.Sprintf("%s.%s", quote(AsString(funcObject.Objname[0])), quote(AsString(funcObject.Objname[1]))),
		}, nil

	case pg_query.ObjectType_OBJECT_COLUMN:
		targetName := commentStmt.GetObject().GetList()
		return &ManagedObject{
			ObjectSchema: AsString(targetName.Items[0]),
			ObjectType:   "comment on column",
			ObjectName:   fmt.Sprintf("%s.%s.%s", quote(AsString(targetName.Items[0])), quote(AsString(targetName.Items[1])), quote(AsString(targetName.Items[2]))),
		}, nil

	case pg_query.ObjectType_OBJECT_VIEW:
		targetName := commentStmt.GetObject().GetList()
		return &ManagedObject{
			ObjectSchema: AsString(targetName.Items[0]),
			ObjectType:   "comment on view",
			ObjectName:   fmt.Sprintf("%s.%s", quote(AsString(targetName.Items[0])), quote(AsString(targetName.Items[1]))),
		}, nil

	default:
		return nil, PKGErrorf(s, nil, "Only comments on views, columns and functions are supported in MOBs")
	}
}

// GetManagedObject returns identifying information about an object from a CREATE
// statement, such as function, view or trigger. NOTE: This function
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

	case stmt.GetCommentStmt() != nil:
		s.object, err = s.getCommentObject()

	case stmt.GetCreateCastStmt() != nil:
		s.object, err = s.getCastObject()

	default:
		clip := strings.TrimSpace(strings.Replace(s.Source[:20], "\n", " ", -1))
		return nil, fmt.Errorf("%s: unknown statement type: %s", s.Location(), clip)
		//fmt.Fprintf(os.Stderr, "WARNING: %s: unknown statement type (in '%s...'); will attempt to execute anyway\n", s.Location(), clip)
		//s.object = &ManagedObject{
		//	ObjectSchema: "unknown",
		//	ObjectType:   "unknown",
		//	ObjectName:   s.Location(),
		//}
	}

	if err != nil {
		return nil, err
	}

	if s.object != nil {
		return s.object, nil
	}

	return nil, PKGErrorf(s, nil, "only functions, triggers, views and comments are supported for managed objects")
}
