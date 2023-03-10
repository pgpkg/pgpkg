package pgpkg

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
)

// Rewrite and the statement source to:
// * Add a SECURITY DEFINER
// * Set the search_path to [schema, temp, public]
//
// This makes functions a bit more secure.
func rewrite(stmt *Statement) error {
	parseResult, err := pg_query.Parse(stmt.Source)
	if err != nil {
		return PKGErrorf(stmt, err, "unable to rewrite function")
	}

	createFuncStmt := parseResult.Stmts[0].Stmt.GetCreateFunctionStmt()

	// FIXME: fire an error if search_path or SECURITY DEFINER is already set.

	createFuncStmt.Options = append(createFuncStmt.Options, getSecurityDefinerOption(), getSetSchemaOption(stmt.Unit.Bundle.Package.SchemaName))

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

// Set search path for all functions in the package.
// See https://www.postgresql.org/docs/current/sql-createfunction.html#SQL-CREATEFUNCTION-SECURITY
func getSetSchemaOption(schema string) *pg_query.Node {
	return &pg_query.Node{
		Node: &pg_query.Node_DefElem{
			DefElem: &pg_query.DefElem{
				Defname: "set",
				Arg: &pg_query.Node{
					Node: &pg_query.Node_VariableSetStmt{
						VariableSetStmt: &pg_query.VariableSetStmt{
							Kind: pg_query.VariableSetKind_VAR_SET_VALUE,
							Name: "search_path",
							Args: []*pg_query.Node{
								&pg_query.Node{
									Node: &pg_query.Node_AConst{
										AConst: &pg_query.A_Const{
											Val: &pg_query.A_Const_Sval{&pg_query.String{Sval: schema}},
										},
									},
								},
								&pg_query.Node{
									Node: &pg_query.Node_AConst{
										AConst: &pg_query.A_Const{
											Val: &pg_query.A_Const_Sval{&pg_query.String{Sval: "pg_temp"}},
										},
									},
								},
								&pg_query.Node{
									Node: &pg_query.Node_AConst{
										AConst: &pg_query.A_Const{
											Val: &pg_query.A_Const_Sval{&pg_query.String{Sval: "public"}},
										},
									},
								},
							},
							IsLocal: false,
						},
					},
				},
			},
		},
	}
}
