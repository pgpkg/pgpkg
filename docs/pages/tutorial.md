# Getting Started with pgpkg

> This tutorial is a work in progress.

## Introduction

pgpkg is a small and fast command-line tool (and Go library) which is designed to make writing Postgresql
stored functions as easy as writing functions in any other language, such as Go, Java or Python.

pgpkg lets you specify function, view and trigger definitions just once. There is no need to
create a migration script every time you change a function, view or trigger.

pgpkg also includes a simple, safe and fast database migration system, and a system for
writing SQL unit tests.

Each of these features will be demonstrated in this tutorial.

## Prerequisites

You need to be able to access a Postgresql database, and you need permission to be able to create
schemas and roles.

> NOTE: `pgpkg` uses the same `PGDATABASE` and other environment variables as `psql`.
> It does not yet support command-line options to override the environment.

You currently need [Go 1.20](https://go.dev/dl/) installed (we will release binary packages soon).

## Permissions and Environment

pgpkg is designed to work with reduced permissions, which may be necessary for hosted
Postgres databases such as Supabase or Vultr. If you have a superuser account, you can
skip this section.

To create a database user `pkgadm` with sufficient privileges to install a package:

    create role pkgadm with createrole login password {{password-in-single-quotes}};
    grant create on database {{PGDATABASE}} to pkgadm;

Then before running pgpkg, you need to set the PG environment:

    export PGUSER=pkgadm
    export PGPASSWORD={{password-in-single-quotes}}

`pgpkg` will automatically create roles for each package that it installs. These roles
will be prefixed with `$` so that they can be easily differentiated from regular roles.
Note however that if you use `$` in your own role names, collisions may be possible.

## Installing pgpkg

Install pgpkg:

    $ go install github.com/pgpkg/cmd/pgpkg

## Writing your first function

Create a folder for your project:

    $ mkdir hello-pgpkg
    $ cd hello-pgpkg

Each pgpkg package requires a small configuration file.  Create one called `pgpkg.toml` file
in the folder project folder:

    Package = "github.com/example/hello-pgpkg"
    Schema = "hello"

Create your first stored function in `func.sql`:

    create or replace function hello.func() returns void language plpgsql as $$
      begin
        raise notice 'Hello, world!';
      end;
    $$;

Note that pgpkg uses schemas extensively; all database objects need to be qualified with
a schema name. We told pgpkg that our schema name is 'hello' in the `pgpkg.toml` file we
just created, so that's what we should use.

With these two files, our `hello-pgpkg` folder now contains a pgpkg package. We can apply
it to the database with a single command:

    $ pgpkg .

(if you want to see what `pgpkg` actually does, use `pgpkg -verbose .`)

If all goes well, you will now have a function defined in your database:

    $ psql
    psql> select hello.func();
    NOTICE:  Hello, world!
    func
    ------
    
    (1 row)

Hmm, that didn't quite work how we wanted. Let's fix that bug!

With traditional migration tools, you would need to add a new version of the function.
With `pgpkg` you just need to edit the existing definition. So, edit func.sql:

    create or replace function hello.func() returns text language plpgsql as $$
      begin
        return 'Hello, world!';
      end;
    $$;

Apply the changes to the database:

    $ pgpkg .

And run it again:

    $ psql
    psql> select hello.func();
    func
    ---------------
    Hello, world!
    (1 row)

That's it! You've written your first pgpkg application without writing a single migration
script.

In `pgpkg`, the function `hello.func` is called a _managed object_ (sometimes abbreviated
to _MOB_). Managed objects don't need migration scripts; you can treat them just like
you would any other code.

This is the main benefit of `pgpkg`: it makes working with functions, views and triggers
much easier.

## Creating a database table

Database tables are _unmanaged objects_, which means they still need to be created and updated
using traditional migration scripts. Let's create one. 

First, create a directory to hold your migration scripts. By convention, we
call this `schema`:

    $ mkdir schema

Let's create a table called 'contact'. Edit `schema/contact.sql`:

    create table hello.contact (
        name text
    );

We need to tell pgpkg the order in which migration scripts need to be run.
To do this, edit the file `schema/@migration.pgpkg`, and add the single line:

    contact.sql

`pgpkg` keeps track of the migration scripts it has already run, so you can simply 
apply the updated package again:

    $ pgpkg .

Let's see if the table exists:

    $ psql
    psql> select * from hello.contact;
    name
    ------
    (0 rows)
    
We forgot to populate the table! So, let's add another migration script.
Call it contact@001.sql, because it's the first change to the contact table. Edit the
file `schema/contact@001.sql`:

    insert into hello.contact (name) values ('Postgresql Community');

Remember that pgpkg needs to know the order in which migrations will run, so you
need to add this new migration script to `schema/@migration.pgpkg`. It should now
look like this:

    contact.sql
    contact@001.sql

You can again apply the updated package to the database:

    $ pgpkg .

Let's see if the data has been added:

    $ psql
    psql> select * from hello.contact;
    name
    ----------------------
    Postgresql Community
    (1 row)

Great! Note that the filename `contact@001.sql` is just a convention. It's not
required by pgpkg, which only cares about the list of filenames in `@migration.pgpkg`.
However, this naming convention means that most IDEs will list migrations in
order, with `contact.sql` followed by `contact@001.sql`. This makes understanding migrations
much easier, especially when there are many of them.

Now, let's use that data in a new function!

Edit the new file `world.sql`:

    create or replace function hello.world() returns text language plpgsql as $$
        declare
            _who text;
    
        begin
            select name into _who strict from hello.contact;
            return _who;
        end;
    $$;
    
Apply the updated package again:

    $ pgpkg .

And now let's see if it worked:

    $ psql
    psql> select hello.world();
            world         
    ----------------------
    Postgresql Community
    (1 row)

It worked! Now, let's write a test to make sure it keeps working.

## Unit Tests

`pgpkg` regards any SQL file ending in `_test.sql` as a test (it doesn't look for
tests in the migration directory, though). Try adding this script to `world_test.sql`
in your project:

    create or replace function hello.test_world() returns void language plpgsql as $$
        begin
            if hello.world() <> 'Postgresql Community' then
                raise exception 'the world is not right';
            end if;
        end;
    $$;

As usual, apply the changes to the database:

    $ pgpkg .

The test will have been applied, but you won't see anything if it passes.

To demonstrate this, use `-summary`:

    $ pgpkg -summary .
    github.com/bookwork/pgpkg: installed 0 function(s), 0 view(s) and 0 trigger(s). 0 migration(s) needed. 0 test(s) run
    github.com/example/hello-pgpkg: installed 2 function(s), 0 view(s) and 0 trigger(s). 0 migration(s) needed. 1 test(s) run

You can see that one test ran in your package (the other package is `pgpkg` itself).

You can add `raise notice` commands to your tests to log information to the console
during the testing process. Edit `world_test.sql` to add a notice:

    create or replace function hello.test_world() returns void language plpgsql as $$
        begin
            raise notice 'Testing the world';
            if hello.world() <> 'Postgresql Community' then
                raise exception 'the world is not right';
            end if;
        end;
    $$;

Let's install it, which will run the test:

    $ pgpkg .           
    [notice]: Testing the world

Any `raise notice` commands from tests that run will be printed to the console.
Messages raised by migrations will also be displayed.

A successful test is one that finds a problem - so let's create a problem!

    $ psql
    psql> update hello.contact set name = 'World';
    UPDATE 1

Now, reinstall the package, which will re-run the tests:

    $ pgpkg .         
    [notice]: Testing the world
    ./world_test.sql:1: test failed: hello.test_world(): pq: the world is not right
           3:         raise notice 'Testing the world';
           4:         if hello.world() <> 'Postgresql Community' then
    -->    5:             raise exception 'the world is not right';
           6:         end if;
           7:     end;
    PL/pgSQL function test_world() line 5 at RAISE

`pgpkg` reports that the test failed - and shows where it happened.

## Package Overview

Here's the directory tree that we created:

    ├── func.sql
    ├── pgpkg.toml
    ├── schema
    │   ├── @migration.pgpkg
    │   ├── contact.sql
    │   └── contact@001.sql
    ├── world.sql
    └── world_test.sql

`world.sql` and `func.sql` contain your stored functions; the `schema` directory
contains your migration scripts.

In just a few files we've been able to create a complete environment for
editing stored procedures in an IDE, the same way we edit regular programming code.
We've also added unit tests - which are just functions that are run after a migration.

## Working with other languages

`pgpkg` will work with any directory that contains `pgpkg.toml`, and will ignore
files that don't end in `.sql`. This means you can mix SQL and (say) Go files in the
same directory:

    ├── func.sql
    ├── go.mod
    ├── go.sum
    ├── main.go
    ├── pgpkg.toml
    ├── schema
    │   ├── @migration.pgpkg
    │   ├── contact.sql
    │   └── contact@001.sql
    ├── world.go
    ├── world.sql
    └── world_test.go

In this example, you can imagine the code in `world.go` is used to access the
function in `world.sql`. You can easily jump between the SQL and Go code in your IDE.
This is the real benefit of pgpkg.

Note that, apart from the schema folder, this directory structure is mostly arbitrary.
You can put tests, functions, views and triggers in any file, as long as there is a
`pgpkg.toml` file in the parent somewhere.

## Integrating with Go

(to be done)

## Purging schemas

When working with a new database schema, you will often want to throw your database
away and start again - you don't want to create a bunch of migration scripts for
changes that nobody will ever see.

pgpkg doesn't currently have a mechanism to enable this, but you can safely drop
the `pgpkg` schema (as well as your other schemas) in order to reset the database.
Take care that any test data you create can be recreated later.

Future versions of `pgpkg` will provide tools to help develop new schemas. For now,
the process is a bit manual.