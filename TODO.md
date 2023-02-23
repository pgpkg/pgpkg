# pgpkg todo

- [ ] loadBundle doesn't support nested subdirectories.
- [ ] make it possible to download/include multiple packages in a project. cli tools? dir format?
- [ ] when tests/table-ref/schema/ref.sql fails, the context is technically correct but visually stupid. 
- [ ] when printing a stack trace (error context), only show the context source for the current package
      e.g. if a test fails when it calls some other package, show the source code location in the test package
      this would make assertions in the pgpkg package (like, =!) work well too.
- [ ] line number in error location headers is wrong (line number doesn't come from context)
- [ ] add support for stored *procedures*
- [ ] delete old code from github
- [ ] make sure only one package can use a schema name at a time (package registration table)
- [ ] not all function parameter types are implemented yet in name generation, e.g. setof. need tests for that. check pgsql syntax too.
- [ ] load bundles in any order, anything not in schema or tests is
      added to applications, which will allow mixing SQL with Go code, big win!
- [ ] generate Go stubs, maybe even Java stubs :-)
- [ ] create a role for schemas so (in theory?) they can't escape the sandbox
- [ ] what happens with quoted identifiers? what happens if the declared schema name is invalid?

# Done

- [X] needs pgpkg.toml - to define the schema
- [X] create the schema if not exists
- [X] make it work for pgpkg itself (so we get the state and other files)
- [X] get rid of "location", make it a function
- [X] add upgrade lock on pgpkg
- [X] add a ; to the end of each script if it doesn't obviously have one
- [X] pgpkg should be failing on item.sql (create trigger) but it's not
- [X] add support for triggers ... once we know why it isn't failing
- [X] need to drop unused functions (just drop them all for now)
- [X] support for views and functions
- [X] parseAll() calls GetObject for each statment; and then so does updateState. Can we dedupe this?
- [X] rename 'application'/'app' to 'api'
- [X] rename "catalog.pgpkg" to "index.pgpkg"
- [X] api.Parse() should be called before schema applied (during load)
- [X] array modifier isn't included when generating function names. 
- [X] get it working as a CLI
- [X] "pgpkg" folder is redundant. just needs a pgpkg.toml file
- [X] move to commandquery
- [X] upgrade to go 20
- [X] getContext() needs to return the statement context if none is found in an error.
- [X] runtime context should return all the contexts up the stack
- [X] move where.go functions into statement.go
- [X] statement.Exec doesn't record the line number or position of errors
- [X] nice error logging for parser errors
- [X] nice error logging for test failures
- [X] implement tests
- [X] tests should be function definitions in the form test_XXX() and called in order
- [X] s.Exec() isn't sufficent, all statements need to be in savepoints
