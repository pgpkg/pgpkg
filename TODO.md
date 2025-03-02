# pgpkg todo

## Priority

- [ ] add a URL in the deprecation warning for @migrations.pgpkg pointing to the FAQ
- [ ] add pointers from the FAQ to the spec and manpage.
- [ ] remove @migration.pgpkg (some time after docs updated) 
- [ ] unmatched $$ at EOL causes a panic
- [ ] don't run tests if nothing's changed
- [ ] don't create .pgpkg unless we're going to put something in it.
- [ ] would be awesome if the REPL option could (perhaps optionally) preserve the test functions. 
- [ ] @migration.pgpkg should just be a list of scripts in pgpgk.toml (which means they can be anywhere, and there's even less config?) - could be part of the TOML even
- [ ] pgpkg.Open() should take a DSN. Only if the DSN is not supplied, use the env.  
- [ ] packages should be purged and installed in (reverse-)dependency order (probably start at project.go: 83 - currently remove and add is done at same time)
  - [ ] need a unit test of function dependencies that fails when dependency order is not honoured

## Testing

- [ ] make sure repl|try|deploy works properly for dependencies, specific path, current directory, child directory, ZIP files
- [ ] check/review that docs still work with (eg, tutorial, manual) - especially after the verb changes
- [ ] create a test for good dependencies, circular dependencies, and missing dependencies.
- [ ] update documentation for toml Uses:, Extensions clause.
- [ ] review & update tutorial & docs to latest standards
  - [ ] add a readme to the example package linking to the tutorial and explaining it a bit.
- [ ] create some structured tests e.g. dependencies, uses, unit tests

## Bugs

- [ ] --[in|ex]clude-tests skips tests in other schemas, should only skip top level package tests (or: should include schema name in patterns)
- [ ] occasional error "unable to drop REPL database pgpkg.xxxxxxxx: unable to drop temp database "pgpkg.isovixpo" when trying `pgpkg repl`
- [ ] the assertion operators don't do anything with null (ie, perform null =? 0 does nothing).
- [ ] change uses of "filepath" to just use "path", ie. filepath.Join() should be path.Join()
- [ ] when a function can't be installed due to an error, and another function depends on it,
  the second function is printed as the error; but the problem is the first function. we should print
  ALL incomplete MOBs if we can't progress, or, at least, the first one to not install.
- [ ] need to remove roles if a package is removed from Uses[]
- [ ] toml Uses[] fails with 'sql: no rows in result set' if a package is not registered. error is ambiguous
- [ ] schema name is missing from function call errors, preventing nice stack traces
- [ ] packages are able to improperly create circular dependencies, which is a security issue, because a dependency
  could trick pgpkg into providing access to a higher level package (not sure if this is still possible; needs checking).
- [ ] when tests/table-ref/schema/ref.sql fails, the context is technically correct but visually stupid.
- [ ] when printing a stack trace (error context), only show the context source for the current package
  e.g. if a test fails when it calls some other package, show the source code location in the test package
  this would make assertions in the pgpkg package (like, =!) work well too.
- [ ] line number in error location headers is wrong (line number doesn't come from context)
- [ ] make sure only one package can use a schema name at a time (package registration table)
- [ ] not all function parameter types are implemented yet in name generation, e.g. setof. need tests for that. check pgsql syntax too.

## Features

- [ ] packages need versioning
- [ ] package up the tool as a binary (github actions?)
- [ ] if a schema hasn't changed (functions, migrations etc) then don't make any changes.
- [ ] make "go test" work with pgpkg
- [ ] allow some kind of "init" or "post" script in MOBs.
- [ ] generate Go stubs, maybe even Java stubs :-)
- [ ] add support for stored *procedure* MOBs
- [ ] remove dependency downloading
- [ ] introspect SQL and plpgsql functions for unwanted statements / set role etc.
  - [ ] ensure search_path and `security definer` are not specified in function definitions
  - [ ] ensure that statements being executed aren't equivalent to "commit", "rollback", "savepoint", "release", etc
  - [ ] ensure that statements being executed aren't SET ROLE or RESET ROLE.

## Docs

- [ ] add examples in the tutorial, maybe supabase, vultr, AWS, some other hosted PGs as well
- [ ] maybe a general "installing psql" document

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
- [X] pgpkg recording OUT params when constructing function signature
- [X] schema path stored in pgpkg.migration is not relative to @index.pgpkg so different invocations do different things.
- [X] pgpkg upgrades can happen in a different tx to others. Open/Init/Install are confusing
- [X] prefix roles with "$" to reduce conflicts with real roles
- [X] rename "API" to "MOB"
- [X] clean up install.go (not needed any more)
- [X] change "@index.pgpkg" to "@migration.pgpkg"
- [X] packages should be rooted at the toml file.
- [X] use pkgadm role instead of pgadm in the docs etc
- [X] package loading should start at the pgpkg.toml file.
- [X] update docs to explain the new package layout / structure
- [X] move Bundle methods out of package and into their own file
- [X] a collection of packages is called a project, which:
- [X] create a role for schemas so (in theory?) they can't escape the sandbox
- [X] Allow more complete integration with source trees:
- [X] try it out using mixed sql + go code
- [X] replace sql.Tx with pgpkg.Tx
- [X] remove package.Exec, logQuery and friends
- [X] there are still a few tx.Exec / tx.Query that should be logged via the package instead
- [X] unsanitized input in Package.sql, maybe other places: `p.Exec(tx, fmt.Sprintf('grant usage on schema "%s" to "%s"')...`
- [X] what happens with quoted identifiers? what happens if the declared schema name is invalid?
- [X] loadBundle doesn't support nested subdirectories.
- [X] refactor options, add --dry-run and friends
- [X] can't disable verbose mode after moving logging into Tx :lol:
- [X] use Go logging.
- [X] use --dry-run when running the tests
- [X] clean up example schema (examples/hello, ...)
- [X] include example in the good tests
- [X] integration tests (including failure tests and success tests)
- [X] update docs to reflect function names are XXX_test (test_XXX still OK but deprecated warning)
- [X] make it so that -- prefix is not needed when running CLI.
- [X] make it possible to disable -- options when running as a library.
- [X] need documentation for --include-tests=, --exclude-tests= and --skip-tests
- [X] repl mode
  - [X] drop database on exit from repl mode
  - [X] trap ^C properly during repl mode
- [X] put pgpkg schema first in schema search and put assertion ops into pgpkg schema
  - [X] this may fix stack traces too, but if not - fix them
- [X] make --dry-run the default, and require --commit to commit.
- [X] add verbs:
  - [X] pgpkg deploy (same as --commit)
  - [X] pgpkg repl   (same as --repl)
  - [X] pgpkg try    (same as deploy --dry-run)
  - [X] pgpkg export (new)
  - [X] update docs
- [X] `pgpkg deploy` should be able to deploy packages exported by `pgpkg export`.
- [X] review code for dead comments
- [X] implement a cache for projects. .pgpkg?
- [X] Project.pkgs should be a map of name:Package rather than an aray, could then also detect dupes
- [X] get local dependencies ("Uses") working.
- [X] once all sources are added to a project, reorder them to resolve dependencies before processing.
- [X] pgpkg commands should default to "current" package, ie, search parents for pgpkg.toml
- [X] ignore dot-files when scanning folders
- [X] need ability to add packages to the cache: `pgpkg cache <path>`
- [X] when importing package c, package b (incorrectly) gets a @migration file, but c doesn't. Why don't they both?
- [X] `pgpkg import` needs to re-import any package that's explicitly mentioned
- [X] `pgpkg import` should import dependencies only if they are not already present in cache; err if dependencies not found
- [X] when importing C, which depends on B; if B is in the local cache then don't import from C's cache; also, it's not an error if it's not in C's cache 
- [X] remove references to "packages.pgpkg"
- [X] Update ZIP writer to just copy the *.sql, toml and migration files, along with the .pgpkg
- [X] non-go deployment
  - [X] needs an "export" option
  - [X] needs documentation (manual) for Go and shell
  - [X] zip package needs to include all dependencies / possibly just use an embedded .pgpkg cache?
- [X] repl is not reliably dropping the "temporary" database (maybe when an error occurs during deploy?)
- [X] "pgpkg repl|try|deploy <file.zip>" should work
- [X] Cache needs to use an FS rather than a path, so it can use Zip caches (zip files should only have a search cache)
- [X] `project` needs a way to identify the main project (maybe in the pkg map)
- [X] `pgpkg import` should not be able to import the current package.
- [X] pgpkg import needs to take two positional parameters, ie: `pgpkg import [target] <source>`
- [X] "pgpkg export <path>" ignores <path>, just uses pwd
- [X] `pgpkg export` still writes packages.pgpkg
- [X] 'pgpkg import' should add a package to Uses, if it's not in Uses already and is not replacing an already-cached package (note: this behaviour is already documented)
- [X] Source should implement fs.FS (ie, by adding Open() method), in which case we could remove FS from Package struct.
- [X] clean up source.go from FS refactor.
- [X] check cmd/pgpkg.go for dead code comments, also project.go, cache.go; and a general review.
- [X] passing an invalid filename (eg, "bc.zipexample") doesn't print an error
- [X] move CLI commands into separate files (makes the code easier to understand)
- [X] pgpkg cli should search parents like Git does, implement Uses
- [X] `pgpkg export` doesn't always write all dependencies - and is non-deterministic about it
- [X] TOML package name checking/sanitation (in config.Uses, config.Name)
- [X] pgpkg export (maybe it should be pgpkg zip?) should name the ZIP file after the package.
- [X] Remove security definer from function definitions
- [X] update README, add release notes, clean up docs. Then we can push it.
- [X] support for CREATE CAST.
- [X] support for arbitrary commands in managed objects (this might be removed later)
- [X] use toml file to list migrations, enable mix of migrations and regular code
  - create table (etc) can't be used in MOBs so risk is minimal
- [X] use filename instead of whole path for migrations - enforce no dupes policy
- [X] check that migration config works with imported packages
- [X] update docs re @migration.pgpkg
