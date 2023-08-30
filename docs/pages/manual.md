# pgpkg

pgpkg is a small, fast and safe Postgresql database migration command-line tool (and Go library) focussed on
ergonomics, simplicity, reliability and reusability:

* ergonomics: SQL function source code can be edited in the same source tree as your native code;
* simplicity: SQL code requires no special migration syntax or file names;
* reliability: migrations are transactional and can include SQL unit tests;
* reusability: SQL code can be reused as modules in multiple projects.

pgpkg makes writing Postgresql stored functions as easy as writing functions in any other language, such as Go, Java
or Python. It lets your SQL code live next to your native code, so you can easily switch between languages in your IDE.

pgpkg is designed explicitly to enable you to use the same source code workflows as your native code: you can
edit your SQL functions in the same IDE, commit them to the same Git repository, review them with PRs alongside other
changes, and deploy them seamlessly to production.

## Usage

    pgpkg {deploy | repl | try | export} [options] [packages]

## Status

pgpkg is alpha quality. It is very usable, but some features aren't yet complete, or work with caveats. In particular,
error reporting is not as good as it could be, and is an area of focus at this time.

## Running pgpkg

The easiest way to run `pgpkg` is:

    pgpkg deploy .

this will search for a `pgpkg.toml` file in the current and parent directories, and perform
a database migration from that point.

`pgpkg deploy` will modify your database. If you want to do a test migration, use `pgpkg try` instead:

    pgpkg try .

`pgpkg try` is identical to `pgpkg deploy`, except that the database transaction used to upgrade the
database is aborted, meaning that your database is not changed.

See [commands](#commands) for more detail.

## Database Connection

pgpkg uses the standard libpq environment variables and defaults (`PGHOST`, etc). If you can run
`psql` with no options, you can use pgpkg. pgpkg also supports a `DSN` environment variable to specify a Postgresql
data source name, which can be used to override the `PG*` settings.

The possible values for `DSN` are documented in
[libpq's connection string](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING)
documentation. You can also use
[libpq parameter key words](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS)
in the DSN variable.

## Configuration

A pgpkg package is any directory (and its children) containing a configuration file called `pgpkg.toml`. The
configuration file contains a package ID, a list of database schemas, and an optional list of database extensions
and external packages to be loaded.

One of the benefits of pgpkg is that it lets you put your SQL function code next to your native code. To achieve
this, `pgpkg.toml` should be placed at the root of your source tree; in Go, this would usually be in the same
folder as `go.mod` or `.git`.

With such a configuration, you can put `.sql` function declarations anywhere in your source tree. (To extract
the SQL definitions from your code when you want to deploy a schema, see `pgpkg export`)

The schemas listed in `pgpkg.toml` are automatically created by pgpkg when it starts.

    Package = "github.com/owner/repository"
    Schemas = [ "schema1", "schema2" ]
    Extensions = [ "ext1", "ext2" ]
    Uses = [ "module1", "module2" ]

pgpkg maintains its own private schema, unsurprisingly called `pgpkg`. This is described in more detail below.

### `Package`

`Package` is the name of the package, using Go's [module path format](https://go.dev/ref/mod#module-path).
You do not need to know Go to use pgpkg. All packages must have a name.

### `Schemas`

`Schemas` is a list of one or more schema names for your package. All packages must exist in at least one schema.
The schemas listed in this clause are created automatically by `pgpkg`.

### `Extensions`

`Extensions` is a list of Postgresql database extensions required by your package. These extensions will be installed
automatically by `pgpgk`.

### `Uses`

`Uses` is a list of `pgpkg` package names which this package depends on. Note that `Uses` is currently
experimental, and not recommended for use in production.

## Managed Objects

In pgpkg, **managed objects** are functions, views or triggers which are explicitly tracked by pgpkg, and installed
automatically as part of a migration. All other objects, including data types and tables, are considered
*migrated objects* (see next section).

Managed objects are stored in any number of `.sql` files, either in the directory containing `pgpkg.toml`, or one of
its children. If `pgpkg.toml` is in the root of your source tree, then you can declare a function simply by creating
an `.sql` file in your source tree, and writing code.

During a migration, **any existing managed objects in the database are automatically dropped**. Migration scripts
are then executed in order (see below). Once this is done, the latest version of managed objects are re-installed.
This process is entirely automatic and transactional; if an error occurs at any time, the transaction is rolled back,
and the database remains intact and usable.

pgpkg automatically resolves dependencies between managed objects in the same schema, so you can declare them
in any order and in any file. For example, if a function `f()` depends on a view `v`, the view will be created before
the function, regardless of where `f()` and `v` are declared in the source tree.

Note that pgpkg expects to have complete control over the creation and removal of managed objects. They should not appear in
migration scripts, and they should not generally be created or dropped outside of pgpkg.

> Exceptions to this rule can, of course, be made during development. Since `.sql` files are literally just SQL scripts
> with no special syntax, it's often convenient to reload function declarations from within `psql` itself.
> For example, if you have functions declared in a file called `launch.sql`, you can quickly update the function
> during debugging using `\i launch.sql`. This isn't generally recommended, but can be a useful hack from time to time.

## Migrated Objects

**Migrated objects** are tables, domains, types, rules, roles, and any other database objects that are not
*managed objects* - that is, any database object that is not a function, view or trigger.

Unlike managed objects, pgpkg never automatically creates or drops migrated objects. Instead, their state
is defined by a (possibly large) set of *migration scripts* which are executed in a defined order. 

Migration scripts are simple `.sql` files containing any number of SQL statements. They are stored in a directory
whose root contains a file called `@migration.pgpkg`. This directory must be a child of the directory
containing `pgpkg.toml`.

You can put any SQL statements at all in a migration script, but you should not generally include functions, views
or triggers, since these are *managed objects* (see above). Example statements in migration scripts might include
any or all of:

* `create table ...`
* `alter table drop column ...`
* `create domain ...`
* `create type ...`

The file `@migration.pgpkg` lists the migration scripts in the order that they are to be executed.
`@migration.pgpkg` can also contain blank lines and comments. The "@" in the filename is intended to make
the file appear at the top of a directory when viewed in an IDE.

Here's an example `@migration.pgpkg` for a fictional database:

    story.sql        # create the story table  
    epic.sql         # create an epic table, also adds epic column to story table.

When you need to make changes to a schema, simply create a new SQL file containing the changes, and add it
to the end of `@migration.pgpkg`.

When `pgpkg` is run, it finds any scripts in the migration catalog which have not already been executed,
and executes them. Scripts are always run in the exact order specified in `@migration.pgpkg`.

**Warning**: The relative path name used for a migration script is used to track if it has been run or not.
You should not change the path of a migration script once it has been executed.

Regardless of the number of migration scripts to be run, all scripts are run a single transaction.
If any migration script fails, the entire operation is aborted and the database is left unmodified.

pgpkg will print an error if a file exists in or under migration directory, but is not listed in `@migration.pgpkg`.

Scripts in the migration directory must not declare managed objects.

Migrated objects are expected to be created only in the schemas declared in `pgpkg.toml`. `pgpkg` may refuse to run
scripts which perform operations outside declared schemas, but this is not yet implemented reliability.

An interesting (and intended) consequence of the way pgpkg migrations work is that it makes it easy for teams to
merge migration scripts from git branches into trunk. Unlike migration tools that use filename sequence numbers to
order migrations, pgpkg's approach means that developers working on branches are unlikely to create merge conflicts.

In the event that two schema changes are made by different teams and then merged, there is little chance that they
will be dependent on one another, which means they can be added to `@migration.pgpkg` in any order.

## Tests

pgpkg supports the writing of SQL unit tests, which are declared as SQL functions. Despite the small learning curve,
we find it much more productive to write tests using `pgpkg`, than to run a schema migration and manually test it
using `psql` (however, we do support this workflow during development: see the [REPL option](#repl), below).

SQL unit tests are executed after a migration has fully completed, but before the migration transaction is committed.

If any test fails, the migration is aborted.

Tests are installed and run inside savepoints. Test savepoints are automatically rolled back before the migration is
complete. Test functions are never visible to production code, and any data created or modified by tests
is also removed (or restored). This lets you write isolated tests which create or delete any data from the database,
without polluting it.

Tests are declared in non-migration scripts whose name matches `*_test.sql`. Test files can appear anywhere managed
scripts can appear. A test script can declare any number of functions.

Functions declared in `_test` scripts are never visible to production code.

If a test script declares a function whose name ends in `_test()`, that function is automatically called by pgpkg
before committing a migration. Such test functions may not have arguments. Here is an example of a test
function declaration:

    create or replace function story.add_test() returns void language plpgsql as ...

Tests are run in random order. Each test is run in an isolated savepoint so that data from one test is not
visible to the other tests.

Tests can call any function declared in any `*_test.sql` files, including functions whose name doesn't end in `_test()`.
Functions declared in `*_test.sql` files are only visible to test functions, and are never visible to production code.

Tests fail when an exception occurs. The exception can be triggered either by SQL itself (for example,
if you try to insert a row with a duplicate key), or explicitly by using
[`raise exception`](https://www.postgresql.org/docs/15/plpgsql-errors-and-messages.html#PLPGSQL-STATEMENTS-RAISE).
Tests can also contain `raise notice` messages which are logged to the console. `raise notice` should only be used
during test debugging.

In addition to the [`assert`](https://www.postgresql.org/docs/15/plpgsql-errors-and-messages.html#PLPGSQL-STATEMENTS-ASSERT)
statement supported by plpgsql, pgpkg declares a number of assertion operators to help write tests. These operators
throw an exception (and fail the test) if the assertion fails. Here are example assertions (all of which pass):

    perform 1 =? 1 and 2 <? 3 and 2 <=? 3 and 3 <=? 3 and 4 >=? 4 and 5 >=? 4 and 6 >? 5 and 6 <>? 7;
    perform 'text' =? 'text' and 'text' <>? 'texta';

There are also two unitary operators, `??` and `?!` for testing predicates:

    perform ??(true), ?!(false);

pgpkg typically runs without printing anything to the console. However, there are
[several pgpkg options for dealing with tests](#transactions-and-testing) which
can print more information about them.

## Transactions

pgpkg always executes within a single database transaction, regardless of the number of migration
scripts which need to be run.

A migration only succeeds if:

- all new scripts listed in `migration.pgpkg` have run without error;
- all managed objects have been successfully recreated; and
- all test functions have run without error.

If any of these statements is not true, or any other error occurs, the migration transaction is aborted
and the database is left unchanged.

Tests are run within savepoints, which are rolled back before completing the migration. Neither test functions,
supporting functions, nor test data are visible to production code after a migration is complete.

## Commands

### `deploy` - deploy packages

    pgpkg deploy [options] <packages...>

`pgpkg deploy` installs the given packages into the database (or updates them if they aren't already present),
and - unless directed otherwise with *options* - runs the SQL unit tests.

If the installation succeeds and the tests pass, `pgpkg deploy` will commit the transaction, resulting in permanent
changes to the database.

If any part of the installation fails, the entire transaction is aborted and the database is left unchanged.

Note that `pgpkg` only applies database migrations that have not already been applied. `pgpkg` will create any
schemas and extensions as needed.

### `try` - test packages

    pgpkg try [options] <packages...>

`pgpkg try` is identical to `pgpkg deploy`, except that, even if deployment is successful, the database transaction is
aborted, and the database is left unchanged. `pgpkg try` lets you try a deployment before committing to it.

### `repl` - interact with packages

    pgpkg repl <packages...>

`pgpkg repl` creates a temporary database, deploys the schema into it (effectively using `pgpkg deploy`), and starts
an interactive `psql`session. This allows you to explore, interact and debug your schema. The temporary database is
dropped when `psql`is exited.

### `export` - create a stand-alone package

    pgpkg export <packages...>

`pgpkg export` creates a single ZIP file containing all the listed packages, and any dependencies.
The resulting ZIP file can be used with `pgpkg deploy` (or `pgpkg try`) to deploy the given packages.

Because `pgpkg` is designed to allow you to mix your native source code and SQL source code, `pgpkg export` provides
a way of extracting only the pgpkg-related files from a source code tree. It is intended for use during an
automated build process; the resulting ZIP file can be shipped with your application to your production
environment, where it can be processed using`pgpkg` as part of the upgrade or application startup process.

For example, in a Java environment, you would use `pgpkg export` to create a `pgpkg.zip` file which you might
bundle into a container along with your application JAR, the JDK, and the `pgpkg` binary. As part of the
container startup script, before starting the Java application, you can run `pgpkg deploy pgpkg.zip`.
This will ensure that the schema is upgraded before the application starts.

> Note: if you are using Go, `pgpkg` supports embedding package files into the Go binary, thereby
> creating a completely stand-alone schema upgrade mechanism that does not require shipping of
> either the schema files or the pgpkg binary. See
> [the pgpkg tutorial](tutorial/go.md) for more information.

## Options

`pgpkg` supports a number of command-line options.

### Testing

`--show-tests`: This option prints a pass/fail status for each test that's run.

`--skip-tests`: do not run any tests before committing the changes. You should take care with this option. 

`--include-tests=[regexp]`: only run tests whose SQL function name matches the given regexp.  

`--exclude-tests=[regexp]`: run all tests, except those whose SQL function name matches the given regexp.

### Logging

pgpkg normally runs silently (unless your SQL code includes `raise notice` messages). These options tell pgpkg
to display more information:

`--verbose`: This option print logs describing *exactly* what pgpkg is up to.

`--summary`: This option print a summary of pgpkg operations when it finishes.

## pgpkg schema

`pgpkg` creates a schema (called `pgpkg`), which contains three tables:

* `pgpkg.pkg`: list of packages that have been installed into this database.
* `pgpkg.managed_object`: list of managed objects that have been installed into this database.
* `pgpkg.migration`: list of migration scripts which have been installed into this database.

These tables should be considered private to `pgpkg`, and the format may change as `pgpkg` evolves.

The schema is, of course, managed using `pgpkg` migrations.

## Parsing

`pgpkg` uses [pganalyse](https://github.com/pganalyze/pg_query_go) to parse SQL scripts before executing them.
The SQL features supported by `pgpkg` are limited to the Postgresql version supported by `pganalyse`.

Most Postgresql features are supported by `pgpkg`, but there may be edge cases which are not yet supported
while in alpha state.

## Examples

See [pgpkg tutorial](./tutorial).

## Installing pgpkg

See [installing pgpkg](./installing.md)

