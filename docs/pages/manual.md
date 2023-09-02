# pgpkg

Packaging tool for Postgresql.

## Usage

    pgpkg {deploy | repl | try | export | import} [options] [packages]

## Description

pgpkg is a small, fast and safe Postgresql database migration command-line tool (and Go library) that
makes it easy to incorporate Postgresql database functions into your native code. It aims to be:

* ergonomic: SQL function source code can be edited in the same source tree as your native code;
* simple: SQL code requires no special migration syntax or file names, and pgpkg has minimal configuration;
* reliable: migrations are transactional - and can include SQL unit tests;
* composable: SQL code can be reused as libraries in other SQL projects.

With pgpkg, you can use the same source code workflows as your native (non-SQL) code: you can
edit your SQL functions in the same IDE, commit them to the same Git repository, review them with PRs alongside other
changes, deploy them seamlessly to production, and even write unit tests.

pgpkg includes tools to help you import other packages into your workspace, deploy your code against a
temporary database, migrate an existing database, run unit tests, and export the SQL code for deployment
into production.

## Status and Compatibility

pgpkg is alpha quality. It is very usable, but some features aren't yet complete, work with caveats, or are buggy.
In particular, error reporting is not as good as it could be, and is an area of focus at this time.

The format for packages, configurations, environment variables, tests etc remains subject to change
until pgpkg 1.0 is released.

pgpkg is not likely to damage your database (unless you write SQL that tells it to), but as with any database tool,
you should always have a backup of your important data, especially when you're getting started.

pgpgk should run on any modern postgresql database. It can be safely used on databases containing
existing tables or functions, and won't modify those tables or functions unless instructed to do so.

pgpkg has not been tested on Windows, and I suspect it won't work due to path issues. This is not
intentional, and I welcome any help in this area.

## Pgpkg packages

A pgpkg package is any directory or ZIP file whose root contains a package configuration in a file called `pgpkg.toml`.

Packages are always installed into one or more database schemas, which must be specified in the
package configuration. pgpkg will create the schemas if they don't exist.

pgpkg packages may also contain a directory called `.pgpkg`. This directory is a cache of packages imported
from elsewhere and used as dependencies in your project.

In pgpkg, SQL functions, views and triggers can be declared in any `.sql` file in the directory tree, meaning
that your SQL functions can be intermingled with your native code. Other database objects, such as tables, domains
and UDTs, are stored in a special directory and are migrated sequentially. The structure of a package is
described in more detail, later.

## Running pgpkg

> Note: by default, `pgpkg` will search for a `pgpkg.toml` file by starting in your working directory and searching
> the parents. In the following examples, we assume you're in a subdirectory of your package directory, but you can
> also specify a project directory.

To create a temporary database, install your package(s), run your tests, and then run `psql`
on the new database:

    pgpkg repl

`repl` is a great tool for quickly testing and exploring your code while developing a new package.
The database is dropped when you quit `psql`.

To perform a test installation of your package into an existing database (specified by the `PG*` environment
variables): 

    pgpkg try

Even if `try` is successful, the database transaction used to upgrade the database is aborted,
meaning that your database is not changed.

To perform a permanent installation of your package into an existing database:

    pgpkg deploy

`deploy` will replace all functions, views and triggers previously installed with `pgpgk` with the latest versions,
and will run a sequential migration of tables and other objects. Note that database objects created outside of
`pgpkg` won't be modified unless explicitly instructed in your `.sql` files.

`pgpkg deploy` is the only one of these commands which will modify an existing database.

See [commands](#commands) for more detail.

## Database Connection

pgpkg uses the standard libpq environment variables and defaults (`PGHOST`, etc), which can be found
[here](https://www.postgresql.org/docs/current/libpq-envars.html). If you can run
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

> **Warning**: `pgpkg` may make changes to a project's TOML file (for example, see `pgpgk import`);
> any comments in the project file will be lost.

A great feature of pgpkg is that it lets you put your SQL function code next to your native code. To achieve
this, `pgpkg.toml` should be placed at the root of your source tree; in Go, this would usually be in the same
folder as `go.mod` or `.git`.

With such a configuration, you can put `.sql` files containing SQL function declarations *anywhere*
in your source tree.

The schemas listed in `pgpkg.toml` are automatically created by pgpkg when it starts.

Here's an example `pgpkg.toml`:

    Package = "github.com/owner/repository"
    Schemas = [ "schema1", "schema2" ]
    Extensions = [ "pgcrypto", "tablefunc" ]
    Uses = [ "github.com/owner/types", "github.com/owner/constants" ]

pgpkg maintains its own private schema, unsurprisingly called `pgpkg`. This is described in more detail below.

### `Package`

`Package` is the name of the package, which consists of an optional domain name followed by an arbitrary path.
`pgpkg` uses a package naming convention similar to Go's module naming convention. Any unique path will work.

### `Schemas`

`Schemas` is a list of one or more schema names that your package will be installed into. pgpkg requires all your
code and tables to be declared only in these schemas. The schemas listed in this clause are created automatically
by `pgpkg`.

### `Extensions`

`Extensions` is a list of Postgresql database extensions required by your package. These extensions will be installed
automatically by `pgpgk` if they don't already exist. `pgpkg` reduces privileges when installing your code; the
`extensions` are installed using the privileges of the invoking user.

### `Uses`

`Uses` is a list of `pgpkg` package names which this package depends on. You can add dependencies to a package using
`pgpkg import <path>` (where <path> is the path to the package you want to import), which will automatically add the
imported package name to the `Uses` clause.

## Functions, Views and Triggers

In pgpkg, functions, views or triggers are called **managed objects**. These objects are declared only once,
in any `.sql` file in your tree. They are explicitly tracked by pgpkg, and installed or upgraded automatically as part
of the deployment process.

All other objects, such as user data types and tables, are considered *migrated objects* (see next section).

Managed objects are stored in any number of `.sql` files, either in the directory containing `pgpkg.toml`, or one of
its children. If `pgpkg.toml` is in the root of your source tree, then you can declare a function simply by creating
an `.sql` file in your source tree, and writing regular SQL code. `pgpkg` does not require any special syntax
for `.sql` files.

During a migration, **any existing managed objects in the database are automatically dropped**. Migration scripts
are then executed in order (see below). Once this is done, the latest version of managed objects are re-installed.
This process is entirely automatic and transactional; if an error occurs at any time, the transaction is rolled back,
and the database remains intact and usable.

pgpkg automatically resolves dependencies between managed objects in the same schema, so you can declare them
in any order and in any file. For example, if a function `f()` depends on a view `v`, the view will be created before
the function, regardless of where `f()` and `v` are declared in the source tree.

Note that pgpkg expects to have complete control over the creation and removal of managed objects. They should not
appear in migration scripts, and they should not generally be created or dropped outside pgpkg.

> Exceptions to this rule can, of course, be made during development. Since `.sql` files are literally just SQL scripts
> with no special syntax, it's often convenient to reload function declarations from within `psql` itself.
> For example, if you have functions declared in a file called `launch.sql`, you can quickly update the function
> during debugging using `\i launch.sql`. This isn't generally recommended, but can be a useful hack from time to time.

## Migrated Objects

**Migrated objects** are tables, domains, types, rules, roles, and any other database objects that are not
*managed objects* - that is, any database object that is not a function, view or trigger.

Unlike managed objects, pgpkg never automatically creates or drops migrated objects. Instead, their state
is defined by a sequence of *migration scripts* - SQL scripts with no special syntax or filenames - which are
executed in the order specified by a migration configuration file. 

Migration scripts are stored in a directory whose root contains a file called `@migration.pgpkg`.
This directory must be a child of the directory containing `pgpkg.toml`. The scripts themselves are
simple `.sql` files containing any number of SQL statements. 

You can put any SQL statements at all in a migration script, but you should not generally include functions, views
or triggers, since these are *managed objects* (see above). Example statements in migration scripts might include
any or all of:

* `create table ...`
* `alter table drop column ...`
* `create domain ...`
* `create type ...`

The file `@migration.pgpkg` lists the migration scripts in the order that they are to be executed.
`@migration.pgpkg` can also contain blank lines and comments.

> The "@" in the filename is intended to make the file appear at the top of a directory when viewed in an IDE.

Here's an example `@migration.pgpkg` for a fictional database:

    story.sql        # create the story table  
    epic.sql         # create an epic table, also adds epic column to story table.

When you need to make changes to a schema, simply create a new SQL file containing the changes, and add it
to the end of `@migration.pgpkg`.

> We've found that using a versioning convention on your migration scripts makes it easier to reason about your schema.
> For example, if you need to make a change to a `story` table created in `story.sql`,
> put those changes in a file called `story@001.sql`. This will group all changes to the `story` table
> together in your IDE. Note that this is just a convention, and not a requirement of `pgpkg`.
> Remember that the specific order that these scripts are run is specified only in the `@migration.pgpgk` file.

When `pgpkg` is run, it finds any scripts in the migration catalog which have not already been executed,
and executes them. Scripts are always run in the exact order specified in `@migration.pgpkg`.

**Warning**: The relative path name used for each migration script is used to track if it has been run or not.
You should not change the path of a migration script once it has been executed.

Regardless of the number of migration scripts to be run, all scripts are run in a single transaction.
If any migration script fails, the entire operation is aborted and the database is left unmodified.

pgpkg will print an error if a file exists in or under the migration directory, but is not listed in `@migration.pgpkg`.

Scripts in the migration directory must not declare managed objects. Doing so is likely to cause unexpected
behaviour.

Migrated objects are expected to be created only in the schemas declared in `pgpkg.toml`. `pgpkg` may refuse to run
scripts which perform operations outside declared schemas, but note that this is not yet implemented reliability.

An interesting (and intended) consequence of the way pgpkg migrations work is that it makes it easy for teams to
merge migration scripts from git branches into main. pgpkg's approach means that developers working on branches
are unlikely to create migration script merge conflicts.

In the event that two schema changes are made by different teams and then merged, there is little chance that they
will be dependent on one another, which means they can be added to `@migration.pgpkg` in any order. Conflicts will only
arise if teams create identical filenames.

## Tests

pgpkg makes it easy to write SQL unit tests, which are declared as SQL functions. Despite the small learning curve,
we find it much more productive to write tests using `pgpkg`, than to run a schema migration and manually test it
using `psql` (however, we do support this workflow during development: see the [REPL option](#repl), below).

SQL unit tests are executed after a migration has fully completed, but before the migration transaction is committed.

If any test fails, the migration is aborted.

Tests are installed and run inside savepoints. Test savepoints are automatically rolled back before the migration is
complete. Test functions are never visible to production code, and any data created or modified by tests
is also removed (or restored).

This lets you write isolated tests which create or delete any data from the database, without polluting it.

Tests are declared in non-migration scripts whose name matches `*_test.sql`. Test files can appear anywhere managed
scripts can appear. A test script can declare any number of functions.

Functions declared in `_test` scripts are never visible to production code.

If a test script declares a function whose name ends in `_test()`, that function is automatically called by pgpkg
before committing a migration. Such test functions must not have arguments. Here is an example of a test
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

pgpkg always executes within a single database transaction, regardless of the number of scripts or tests
which need to be run.

A migration only succeeds if:

- all new scripts listed in `migration.pgpkg` have run without error;
- all managed objects have been successfully recreated; and
- all test functions have run without error.

If any of these statements is not true, or any other error occurs, the migration transaction is aborted
and the database is left unchanged.

Tests are run within savepoints, which are rolled back before completing the migration. None of the test functions,
test-supporting functions, inserted, updated or deleted data are visible to production code after a migration is
complete.

## Commands

### `deploy` - deploy packages

    pgpkg deploy [options] [package]

`pgpkg deploy` installs a package into the database (or updates them if they are already present),
and - unless directed otherwise with *options* - runs the SQL unit tests.

If `package` is specified, it should name a directory or ZIP file containing the package to be installed.
If not specified, `pgpkg` will search the current working directory and parent directories to find a `pgpkg.toml`
file, and install from there. This allows the command to be run from any subdirectory of a package.

If any `options` are specified, they are processed as described below.

If the installation succeeds and the tests pass, `pgpkg deploy` will commit the transaction, resulting in permanent
changes to the database.

If any part of the installation fails, the entire transaction is aborted and the database is left unchanged.

For migrated objects, **`pgpkg` only applies database migrations that have not already been applied**. It records
the pathname of each successful migration, and will only apply a given migration once. Migrations which have not yet
been performed are then applied in the order specified.

`pgpkg` will create the schemas and extensions specified in the `pgpkg.toml` file.

### `try` - deployment dry run

    pgpkg try [options] [package]

`pgpkg try` is identical to `pgpkg deploy`, except that, even if deployment is successful, the database transaction is
aborted, and the database is left unchanged. `pgpkg try` lets you try a deployment before committing to it.

The optional `package` argument is documented in `pgpkg deploy`. 

If any `options` are specified, they are processed as described below.

### `repl` - interact with packages

    pgpkg repl [options] [package]

`pgpkg repl` creates a temporary database, deploys the schema into it (effectively using `pgpkg deploy`), and starts
an interactive `psql`session. This allows you to explore, interact and debug your schema. The temporary database is
dropped when `psql`is exited.

The optional `package` argument is documented in `pgpkg deploy`.

If any `options` are specified, they are processed as described below.

### `import` - import a package into the current package

    pgpkg import [package] <from-package>

`pgpgk import` imports `from-package` into the given package (or the default package
if one is not specified), and adds it to the Uses clause of the `pgpkg.toml` file
if it's not there already.

The import process copies `pgpkg` files including `pgpkg.toml`, `@migration.pgpkg`, the directory
structure, and all `.sql` files into a cache folder called `.pgpkg` in the target package. If the SQL code
in the from-package is mixed with native code, the native code is **not** copied.

If the specified `from-package` has dependencies that have not previously been imported into the current package,
those dependencies are also imported.

> **Warning** this command updates a package's `pgpkg.toml` file, which will strip any comments
> out of the file.

This command allows you to include SQL packages as dependencies in the `Uses` clause of your package.
The imported package can create user data types, functions, system views and any other object (including sequentially
migrated database tables) which can then be used by the importing package.

The optional `package` argument, referring to target package into which the from-package is to be imported,
is documented in `pgpkg deploy`.

The required `from-package` is the path to a package to be imported.

`pgpkg import` will always import the specified package into the cache. If necessary, it will also
import missing dependencies, but it will decline to import dependencies that already exist in the cache.
This behaviour is intended to avoid inadvertently modifying packages you already depend on.
To upgrade existing packages, import them directly.

Note that pgpkg does not (currently) implement or support package versioning.

### `export` - export a package into a deployable archive

    pgpkg export [package]

`pgpkg export` creates a single ZIP file containing the given package, including any dependencies.
The resulting ZIP file can be used directly with `deploy`, `repl`, `try` or `import`.

The optional `package` argument is documented in `pgpkg deploy`.

Because `pgpkg` is designed to allow you to mix your native source code and SQL source code, `pgpkg export` provides
a way of extracting only the pgpkg-related files from a source code tree. It is intended for use during an
automated build process; the resulting ZIP file can be shipped with your native application to your production
environment, where it can be processed using `pgpkg` as part of the upgrade or application startup process.

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

## Package Cache

A `pgpkg` package will typically contain a *cache directory*; this cache is stored in a subdirectory called `.pgpkg`
in the top-level package directory.

If a package depends on one or more other packages, those packages *must* exist in the cache. In this way,
a package is generally self-contained. Packages in the cache are stored in a directory tree based on the package
path.

You can add packages to the cache using `pgpkg import`, which will also add the specified package as a dependency
to your project.

ZIP files created by `pgpkg export` will include a `.pgpkg` directory containing dependencies, thereby making
ZIP files self-contained.

> Note that files in `.pgpkg` should be included in source code control.

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

