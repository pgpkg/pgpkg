# pgpkg todo

- [ ] move to commandquery
- [ ] remove old pgpkg from cq and install new version 
- [ ] implement tests
- [ ] nice error logging for parser errors
- [ ] nice error logging for test failures
- [ ] make it possible to download/include multiple packages in a project. cli tools? dir format?
- [ ] not all function parameter types are implemented yet, e.g. setof
- [ ] make sure only one package can use a schema name at a time
- [ ] function return type should be recorded in object state
      this lets us know if we need to drop them or not.
- [ ] we might want to make it so a function can be replaced, e.g.
      this would allow tables to depend on a function or view.
- [ ] load bundles in any order, anything not in schema or tests is
      added to applications, which will allow mixing SQL with Go code.
- [ ] loadBundle doesn't support nested directories.

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
