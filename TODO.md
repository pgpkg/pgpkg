# pgpkg todo

- [ ] pgpkg upgrades can happen in a different tx to others. Open/Init/Install are confusing.
- [ ] packages are treated individually which will cause dependency problems.
- [ ] rename "API" to "Managed" or "Declarations" or something (also in the example package structure)
- [ ] pgpkg cli should search parents like Git does
- [ ] package up the tool as a binary
- [ ] package loading should start at the pgpkg.toml file.
- [ ] update docs to explain the new package layout / structure
- [ ] Allow more complete integration with source trees:
    - [ ] change "@index.pgpkg" to "@migration.pgpkg"
    - [ ] try it out using mixed sql + go code
- [ ] move Bundle methods out of package and into their own file
- [ ] **important** pretty sure we don't delete from the object list in the database after a purge
      results in old object signatures being repeatedly "drop ... if exists"
- [ ] toml Uses[] fails with 'sql: no rows in result set' if a package is not registered. error is ambiguous
- [ ] unsanitized input in Package.sql, maybe other places: `p.Exec(tx, fmt.Sprintf('grant usage on schema "%s" to "%s"')...`
- [ ] schema name is missing from function call errors, preventing nice stack traces
- [ ] loadBundle doesn't support nested subdirectories.
- [ ] ensure search_path and `security definer` are not specified in function definitions
- [ ] ensure that statements being executed aren't equivalent to "commit", "rollback", "savepoint", "release", etc
- [ ] ensure that statements being executed aren't SET ROLE or RESET ROLE.
- [ ] views need security definers too
- [ ] there are still a few tx.Exec / tx.Query that should be logged via the package instead
- [ ] introspect SQL and plpgsql functions for unwanted statements / set role etc.
- [ ] packages are able to improperly create circular dependencies, which is a security issue, because a dependency
      could trick pgpkg into providing access to a higher level package.
- [ ] dependency management, download, registration, etc
- [ ] when tests/table-ref/schema/ref.sql fails, the context is technically correct but visually stupid. 
- [ ] when printing a stack trace (error context), only show the context source for the current package
      e.g. if a test fails when it calls some other package, show the source code location in the test package
      this would make assertions in the pgpkg package (like, =!) work well too.
- [ ] line number in error location headers is wrong (line number doesn't come from context)
- [ ] add api support for stored *procedures*
- [ ] make sure only one package can use a schema name at a time (package registration table)
- [ ] not all function parameter types are implemented yet in name generation, e.g. setof. need tests for that. check pgsql syntax too.
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
- [X] create roles for each package
- [X] add search_path to created functions. (nb: views and triggers will Just Work™)
- [X] support for ZIP files containing schemas.
- [X] instead of tests folder, use xxx_test.sql (similar to go)
- [X] migration should be any folder that contains @index.pgpkg
- [X] walk the filesystem from any root (but handle @migration dirs separately)
- [X] anything *.sql, not in a folder with @migration.sql and not ending in _test is an API.
- [X] Schema.readCatalog() is redundant if loadPackage2 works. (it's called from Schema.Apply())
- [X] reduce logVolume when installing extensions to avoid 'extension "uuid-ossp" already exists"
