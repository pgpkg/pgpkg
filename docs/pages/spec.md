# pgpkg Package Specification

> Note: this page describes the *structure* of pgpkg packages. See [pgpkg phases](phases.md)
> for information about how the installation process actually installs packages into
> your database.

## Introduction

`pgpgk` works by scanning files in a directory (or Go filesystem), determining
the purpose of each SQL file from the context, and executing the SQL file in
a particular way.

`pgpkg` considers any directory that contains a pgpkg.toml file to be a package. 

Packages can contain _managed objects_, _migration scripts_ and _tests_.

The rules for a `pgpkg` package are simple:

* SQL scripts containing test functions must be named `*_test.sql`
* Migration scripts must be stored under a directory containing the file `@index.pgpkg`.
* All other SQL scripts are _managed_, and may contain only function, view, and
  trigger definitions.

`pgpkg` packages are installed into well-defined database schemas. A package is installed into exactly one schema.
`pgpkg` makes an effort to ensure that objects are installed only into the schema that
they declare, but please beware that this checking is, at present, rudimentary (and probably unsafe).
Read more about [package safety with pgpkg here](safety.md).

`pgpkg` only manages objects that it knows about. You can use pgpkg to "take over"
from an existing schema, without worrying that objects defined outside pgpkg will be modified.
You can also use `pgpkg` to take over migration and function definitions for an existing database.
However, when introducing pgpkg into an existing database, take care to ensure that managed definitions
don't conflict with unmanaged definitions.

## Style Guide

Although not enforced or required by `pgpkg`, we recommend that stand-alone Postgresql packages implement
the file structure demonstrated in the [example package](https://github.com/pgpkg/pgpkg-test). For example:

    ├── api             
    │   └── api_1.sql      -- SQL files containing functions, views and triggers
    │   └── ...
    ├── pgpkg.toml         -- package configuration
    ├── schema             -- migration files 
    │   ├── @index.pgpkg   -- ordered list of files to migrate
    │   ├── one.sql 
    │   ├── two.sql        -- SQL files containing migration code
    │   └── ...
    └── tests
        ├── test1_test.sql -- SQL files containing test functions 
        └── ...

The file structure shown above should be used for packages that are intended for inclusion in other projects.

Note, however, that `pgpkg` is also designed to be used alongside a host language such as Go or Java.
For this reason, it is OK to mix SQL and non-SQL files in the same codebase, which is useful if
a package isn't meant for distribution. `pgpkg` will ignore any files it doesn't recognise.

## pgpkg.toml

`pgpkg.toml` is a [TOML](https://toml.org) file describing the package itself. All SQL
files for a package must appear under the directory containing this file.

There is not much to a package configuration:

    # A pgpkg TOML file.
    Package = "github.com/pgpkg/example"
    Schema = "example"
    Extensions = [ "uuid-ossp" ]
    Uses = [ "github.com/pgpkg/another-example" ]

`Package` is a unique package identifier. It's intended to be a URL where the package can
be downloaded.

`Schema` is the schema into which the package will be installed. You can't install two different packages
into the same schema.

`Extensions` is a list of extensions that the package requires. `pgpkg` will attempt to
install these extensions for you.

`Uses` provides a list of packages that this package depends on. pgpkg will automatically
add grants to the package's role to enable it to access these packages.

> pgpkg doesn't yet manage dependencies; you have to install them manually.
> However, pgpkg does keep track of which packages have been installed, and will automatically
> manage upgrades of those packages.

When a package is installed, `pgpkg` creates both the named schema and a matching role. Objects
created in the package are owned by the role. Except when a `Uses` clause is present,
the role created for a package does not have permission to access any other package.

See [safety](safety.md) for more information.

## Migrations

Any directory containing a file called `@index.pgpkg` is considered to be a
migration directory. There can only be a single migration directory in a package.
SQL files can appear directly inside this directory, or inside a subdirectory.

We recommend naming this directory `schema`, but that is not a requirement of `pgpkg`.

Managed objects and tests cannot appear in a migration directory.

Only files named in the `@index.pgpkg` file will be used in a migration. `pgpkg`
will print a warning if a SQL file is found in the migration directory, but is
not listed in `@index.pgpkg`.

The migration folder can contain SQL scripts that are typical for a database migration tool.
When pgpkg begins a migration, it runs any scripts it hasn't seen before, the order specified
in `@index.pgpkg`.

Scripts are run with the schema name as the effective database role, so objects created
by a migration are owned by the package's schema.

The path name of migration scripts, relative to the migration directory, is used to determine
if a migration script needs to be run. This path name includes the name of any subdirectories
under the migration directory. When successful, the name of each script is stored in the
`pgpkg` schema's database, thus preventing the script from running again in the future.

> WARNING: Renaming a migration script may cause the script to run again during a
> migration. Once you deploy a migration, it should not be renamed.

`@index.pgpkg`, which must exist in the top level migration directory, lists the filenames of
the migration scripts that must be run. Migration scripts are run strictly in the order that they
are found in this file. The file may refer to scripts found in subdirectories of the
migration folder.

File names in `pgpkg` do **not** determine the order of execution of a migration.

Using an index file means you can group migration scripts by name, which makes managing
them much easier. For example, if a script called `example.sql` creates a table called `example`,
and we later need to add a column to it, we can do that in a file called `example@001.sql` -
which will appear under the `example.sql` script in most IDEs. However, the new script
can be added to the end of the `@index.pgpkg` file, meaning that it will run last.

Alternatively, we could also put the `create table` and `alter table` scripts in a subdirectory
called `example`. This will let us group all changes for an entity together, which makes it easier to
understand what changes have been made to a single entity over time.

All pgpkg migrations are executed in a single transaction, and if any migration fails,
the entire transaction is automatically rolled back. This provides you with atomic upgrades and
gives you a failsafe if something goes wrong.

Scripts in the migration folder are written in plain SQL. Unlike managed and test scripts,
migration scripts can contain any SQL whatsoever. Migration scripts are intended
to be used to create and update persistent objects such as tables, indexes and data types, and
shouldn't generally be used to implement functions, views or triggers.

## Managed Objects

SQL files that aren't in the migration directory and whose name doesn't end in `_test.sql`
are considered to contain *managed object declarations*.

Managed objects are functions, views and/or triggers which
`pgpgk` installs into your database. SQL files may contain definitions for any number
of managed objects, but they must only create functions, views or triggers.

Managed SQL scripts contain plain, regular SQL. There are no magic patterns, special
comments or anything else. This design is intended to give you flexibility during
development, by enabling you to run your managed SQL scripts using `psql` (or your IDE),
which can speed up the REPL process.

Managed scripts can contain any number of function, view or trigger definitions.
For example, a file called `example.sql` might contain two function definitions:

    create or replace function example.hello() returns void language plpgsql as $$
        begin
            raise notice 'hello, world';
        end;
    $$;

    create or replace function example.world() returns void language plpgsql as $$
        begin
            raise notice 'hello, world';
        end;
    $$;

`pgpkg` will extract the CREATE definitions for each managed SQL object,
and run them.

In order to determine local dependencies, `pgpkg` automatically works out the order in which functions,
views and triggers must be installed; for example, a view that refers to a function will cause the 
function to be created first.

In this way, pgpkg eliminates the need to write migration scripts for functions, views
and triggers. 

Note that, as part of the upgrade process, pgpkg deletes any functions, views or triggers
it manages, before attempting to install the latest versions. Although pgpkg won't modify any
object it didn't create, if you use SQL to modify something pgpkg created, your changes may be
lost when `pgpkg` is run again.

At the time of writing, only the following DDL commands are can be used in managed SQL scripts:

    create or replace function ...;
    create or replace view ...;
    create or replace trigger ...;

pgpkg keeps track of the functions, views and triggers it creates based on your code.

See [pgpkg phases](phases.md) for more information about how the installation process
works.

## Tests

pgpkg lets you specify any number of unit tests, which are run after a migration
is completed. Tests are defined in all files whose name end in `_test.sql`.

Test scripts can contain any number of functions, but only functions whose name
starts with `test_` will be run.

For example, a file named `example_test.sql` might contain the following unit test: 

    create or replace function example.test_world() returns void language plpgsql as $$
        begin
            raise notice 'hello, test!';
        end;
    $$;

Like everything else in pgpkg, tests are just regular SQL scripts. However:

* Tests can only define functions, not view or triggers.
* Functions whose name starts with `test_` are run directly.
* Tests can call other functions.
* Tests are run in random order.

Importantly, tests are executed in a transaction which is **always rolled back**. This
means that the functions created by the test scripts, and any data created or modified by
those tests, is also rolled back. You can therefore write tests which modify the database,
because those modifications will be deleted once the tests are complete.

Tests can log information to the console using `raise notice`, as seen above. These
notices will be printed to the console during the migration process.

To fail a test, use `raise exception`:

    create or replace function example.test_fail() returns void language plpgsql as $$
        begin
            raise exception 'eject!';
        end;
    $$;

Regardless of the success or failure of the tests, they leave no trace once they
complete. Any functions created in a `*_test.sql` file are removed before the
migration completes.