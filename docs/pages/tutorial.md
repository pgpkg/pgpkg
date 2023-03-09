# Getting Started with pgpkg

> This tutorial is a work in progress.

## Introduction

pgpkg reduces the hassle of writing and shipping stored functions for Postgresql.

With pgpkg, there is no need to create a migration script every time you change a stored
function, view or trigger. You just change the original file, and pgpkg looks after the rest.

In addition to managing stored functions, views and triggers separately from other database
objects, pgpkg also includes a safe and fast database migration system, and an SQL
based unit testing tool.

Each of these features will be demonstrated in this tutorial.

## Prerequisites

You need to be able to access a Postgresql database, and you need permission to be able to create
schemas and roles.

> NOTE: `pgpkg` uses the same `PGDATABASE` and other environment variables as `psql`.
> It does not yet support command-line options to override the environment.

You currently need [Go 1.20](https://go.dev/dl/) installed (we will release binary packages soon).

## Permissions and Environment

> If you have a superuser account on your Postgresql instance, you can skip this section.

pgpkg is designed to work with reduced permissions, which may be necessary for hosted
Postgres databases such as Supabase or Vultr. 

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

(if you want to see what `pgpkg` actually does, use `pgpkg -pgpkg-verbose .`)

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

To demonstrate this, use `-pgpkg-summary`:

    $ pgpkg -pgpkg-summary .
    github.com/bookwork/pgpkg: installed 0 function(s), 0 view(s) and 0 trigger(s). 0 migration(s) needed. 0 test(s) run
    github.com/example/hello-pgpkg: installed 2 function(s), 0 view(s) and 0 trigger(s). 0 migration(s) needed. 1 test(s) run

You can see that one test ran in your package (the other package is `pgpkg` itself).

You can add `raise notice` (and `raise warning`) commands to your tests to log information
to the console during the testing process. Edit `world_test.sql` to add a notice:

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
(`raise warning` commands are printed to stderr). Messages raised in migration
scripts will also be displayed.

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

## Integrating with Go

One of pgpkg's features is that it plays nicely with other languages. It plays especially
nicely with Go, but you can use the features of pgpkg with any language.

In this example we're going to integrate our application with a Go program.

First, you need to create a module and add pgpkg as a dependency:

    $ go mod init github.com/pgpkg/example
    go: creating new go.mod: module github.com/pgpkg/example
    go: to add module requirements and sums:
    go mod tidy
    $ go get github.com/pgpkg/pgpkg

Next, we need to set up main.go to load the schema and open the database:

    package main

    import (
        "embed"
        "fmt"
        "github.com/pgpkg/pgpkg"
    )
    
    //go:embed pgpkg.toml *.sql schema
    var hello embed.FS

    func main() {
        pgpkg.ParseArgs()
        var p pgpkg.Project
        p.AddFS(hello)
    
        db, err := p.Open()
        if err != nil {
            pgpkg.Exit(err)
        }
        defer db.Close()
    
        var world string
        if err = db.QueryRow("select hello.world()").Scan(&world); err != nil {
            pgpkg.Exit(err)
        }
    
        fmt.Println("And the world is:", world)
    }

Running this program will do the following:

* parse command line options like "-pgpkg-verbose"
* run any necessary migration
* install the SQL function
* run the tests, and
* return a database handle.

So let's build and run it:

    $ go build
    $ ./example
    [notice]: Testing the world
    And the world is: Postgresql Community

Note how the test has printed a message - `raise notice` commands are automatically
logged from pgpkg. We can fix this by updating the test. Edit world_test.sql to remove
the notice:

    create or replace function hello.test_world() returns void language plpgsql as $$
        begin
            if hello.world() <> 'Postgresql Community' then
                raise exception 'the world is not right';
            end if;
        end;
    $$;

Then build and rerun your code:

    $ go build
    $ ./example
    And the world is: Postgresql Community

> Note that you don't need to run `pgpkg` when you've embedded your package into your Go
> program. The Go program migrates the database for you, when you call p.Open().

`pgpkg.ParseArgs` adds support for a standard set of migration options to your Go program.
These include:

    -pgpkg-summary
    -pgpkg-verbose
    -pgpkg-dry-run

For example, we can run this:

    $ ./example -pgpkg-summary
    github.com/bookwork/pgpkg: installed 0 function(s), 0 view(s) and 0 trigger(s). 0 migration(s) needed. 0 test(s) run
    github.com/example/hello-pgpkg: installed 2 function(s), 0 view(s) and 0 trigger(s). 0 migration(s) needed. 1 test(s) run
    And the world is: Postgresql Community

Finally, note how your SQL and Go code can live in the same directory:

    .
    ├── example
    ├── func.sql
    ├── go.mod
    ├── go.sum
    ├── main.go
    ├── pgpkg.toml
    ├── schema
    │   ├── @migration.pgpkg
    │   ├── contact.sql
    │   └── contact@001.sql
    ├── world.sql
    └── world_test.sql

This makes switching between Go and plpgsql code in your IDE completely seamless.

## Using pgpkg in other languages

To use pgpkg in languages other than Go, you can simply package up the pgpkg CLI command as
part of the startup script for your application. pgpkg can read ZIP files, and the
bundling can occur as part of the deployment build process.

## Purging schemas

When working with a new database schema, you will often want to throw your database
away and start again - you don't want to create a bunch of migration scripts for
changes that nobody will ever see.

pgpkg doesn't currently have a mechanism to enable this, but you can safely drop
the `pgpkg` schema (as well as your other schemas) in order to reset the database.
Take care that any test data you create can be recreated later.

Future versions of `pgpkg` will provide tools to help develop new schemas. For now,
the process is a bit manual.