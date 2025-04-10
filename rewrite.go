package pgpkg

import (
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

// Rewrite the statement source to set the search_path to [schema, temp, public]
// NOTE: rewriting doesn't seem to preserve the quotes in function arguments.
// This appears to be a problem with pg_analyze_go.
func rewrite(stmt *Statement) error {

	parseResult, err := pg_query.Parse(stmt.Source)
	if err != nil {
		return PKGErrorf(stmt, err, "unable to rewrite function")
	}

	createFuncStmt := parseResult.Stmts[0].Stmt.GetCreateFunctionStmt()

	// FIXME: fire an error if search_path or SECURITY DEFINER is already set.

	//schemaNames := append([]string{"pgpkg"}, stmt.Unit.Bundle.Package.SchemaNames...)
	//schemaNames = append([]string{"pg_temp", "public"}...)
	schemaNames := []string{"pgpkg", "pg_temp", "public"}

	createFuncStmt.Options = append(createFuncStmt.Options /* getSecurityDefinerOption(), */, getSetSchemaOption(schemaNames))

	stmt.Source, err = pg_query.Deparse(parseResult)
	if err != nil {
		return PKGErrorf(stmt, err, "unable to generate rewritten function")
	}

	return nil
}

func getSecurityDefinerOption() *pg_query.Node {
	return &pg_query.Node{
		Node: &pg_query.Node_DefElem{
			DefElem: &pg_query.DefElem{
				Defname: "security",
				Arg: &pg_query.Node{
					Node: &pg_query.Node_Boolean{
						Boolean: &pg_query.Boolean{
							Boolval: true,
						},
					},
				},
			},
		},
	}
}

// Set search path for all functions in the package, to the schemas declared for the package.
// This means you don't need to schema-qualify code inside the package, but it's still a good
// idea.
// See https://www.postgresql.org/docs/current/sql-createfunction.html#SQL-CREATEFUNCTION-SECURITY
func getSetSchemaOption(schemaNames []string) *pg_query.Node {
	return &pg_query.Node{
		Node: &pg_query.Node_DefElem{
			DefElem: &pg_query.DefElem{
				Defname: "set",
				Arg: &pg_query.Node{
					Node: &pg_query.Node_VariableSetStmt{
						VariableSetStmt: &pg_query.VariableSetStmt{
							Kind:    pg_query.VariableSetKind_VAR_SET_VALUE,
							Name:    "search_path",
							Args:    getSchemaNameArgs(schemaNames),
							IsLocal: false,
						},
					},
				},
			},
		},
	}
}

func getSchemaNameArgs(schemaNames []string) []*pg_query.Node {
	var nodes []*pg_query.Node

	for _, schema := range schemaNames {
		nodes = append(nodes, &pg_query.Node{
			Node: &pg_query.Node_AConst{
				AConst: &pg_query.A_Const{
					Val: &pg_query.A_Const_Sval{&pg_query.String{Sval: schema}},
				},
			},
		})
	}

	return nodes
}
