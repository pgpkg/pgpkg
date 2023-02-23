# Package Structure

pgpkg takes a directory structure and applies it to a database.

A pgpkg package is simply a directory that looks like this:

    pgpkg.toml
    api/
    schema/
    tests/

The api, schema and tests directories are used in specific ways, which are described
in detail below.

> Note: this page describes the *structure* of the tests. See [pgpkg phases](phases.md)
> for information about how the installation process actually installs this code into
> your database.

## Database Schemas

pgpkg packages are installed into database schemas. A package is installed into exactly one schema.
At this time, pgpkg makes an effort to ensure that objects are installed only into the schema that
they declare, but please beware that this checking is rudimentary and, probably, unsafe.
([Read more](safety.md) about safety with pgpkg).

pgpkg only manages objects that it knows about. For example, you can use pgpkg to "take over"
from an existing schema, without worrying that objects defined outside pgpkg will be modified.

## pgpkg.toml

The pgpkg.toml file describes the package itself. Currently, this file is pretty small:

    # A working test package for pgpkg.
    Package = "github.com/pgpkg/example"
    Schema = "example"

In the future, this file will also list dependencies and may be used to provide more information
about packages.

## Schema Directory

The `schema/` folder contains SQL code that is typical for a migration tool.
It is, effectively, just a set of scripts which are executed sequentially. pgpkg keeps track
of which files have been installed and which haven't. When pgpkg starts, the first thing
it does is run any scripts it hasn't seen before, in a specific order.

The `schema/` folder must include a file called `@index.pgpkg`. This file, in turn, must
list the order in which migration scripts must be run.

> Some migration tools use the filename to determine the order of migrations.
> pgpkg requires the files to be listed explicitly, which means you can group
> migration scripts by entity. For example, if a script called `example.sql`
> creates a table called `example`, and we later need to add a column to it,
> we can do that in a file called `example@001.sql` (we could also put the
> changes in a subdirectory). This lets us group all changes for an entity
> together, which makes it easier to understand what changes have been made
> to an entity over time.

All pgpkg migrations are executed in a single transaction, and if any migration fails,
the entire transaction is rolled back. This provides you with atomic upgrades and
gives you a failsafe if something goes wrong.

Similar to the `api/` folder, scripts in the `schema/` folder are just plain SQL.
Unlike the `api` folder, `schema` scripts can contain any SQL whatsoever. However,
`schema` scripts are intended to be used to create and update persistent objects such
as tables, indexes and data types.

> WARNING: The name of the scripts in `schema` is used to determine if the script has 
> already run. Renaming a script in this folder may cause it to run again.

## API

The `api/` folder contains SQL scripts which create functions, views and triggers. This
code is considered the API for your database, and it's expected that other code you write
will use these objects either directly (in the case of functions and view) or indirectly
(in the case of triggers).

The api folder just contains plain, regular SQL. There are no magic patterns, special
comments or anything else. pgpkg is deliberately designed to allow you to run your SQL
code using `psql` during development, if that makes you go faster.

SQL files in the `api` folder can contain any number of object definitions.
For example, `api/example.sql` might contain two function definitions:

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

pgpkg will read all files under the `api` tree, extract the CREATE definitions from
them, and then run them. pgpkg will work out the order in which functions, views and
triggers need to be installed; for example, a view that refers to a function will be
created first.

In this way, pgpkg eliminates the need to write migration scripts for functions, views
and triggers. 

Importantly, pgpkg will delete any functions, view or triggers it created previously,
before attempting to install the new objects. pgpkg won't delete a function it didn't
create, but note that if you overwrite something pgpkg created, it's fair game.

At the time of writing, the `api` folder only supports the following statements:

    create or replace function ...;
    create or replace view ...;
    create or replace trigger ...;

pgpkg keeps track of the functions, views and triggers it creates based on your code.

See [pgpkg phases](phases.md) for more information about how the installation process
works.

## Tests

pgpkg lets you specify a set of tests which are run after a migration is completed.

Each SQL file in the `tests` folder declares a set of functions. For example:

    create or replace function example.test_world() returns void language plpgsql as $$
        begin
            raise notice 'hello, test!';
        end;
    $$;

Like everything else in pgpkg, tests are just regular SQL scripts. However:

* Tests can only define functions, not view or triggers.
* Functions whose name starts with `test_` are run directly.
* Tests are run in random order.

Importantly, tests are executed in a transaction which is always rolled back. This
means that the functions created by the tests, and any data created or modified by the
tests, is also rolled back.

Tests can log information to the console using `raise notice`.

To fail a test, use `raise exception`:

    create or replace function example.test_fail() returns void language plpgsql as $$
        begin
            raise exception 'eject!';
        end;
    $$;

pgpkg receives `raise notice` messages from the database and logs them to the console.
These notices can come from any function in the database, not just test functions.
