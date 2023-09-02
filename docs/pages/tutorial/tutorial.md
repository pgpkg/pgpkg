# Getting Started with pgpkg

> If you want to jump straight to the code, the final tutorial code can be found
> [on GitHub](https://github.com/pgpkg/pgpkg/tree/main/tests/good/example).

## Introduction

pgpkg is a database migration tool that makes it much easier to write and deploy stored functions, views and
triggers in Postgresql.

With pgpkg, there is no need to create a migration script every time you change a stored
function, view or trigger. You just change the original file, and pgpkg looks after the rest.

In addition to managing functions, views and triggers separately from other database
objects, pgpkg also includes a **safe and fast sequential migration system** for tables and other objects,
and an SQL based **unit testing tool**.

Each of these features will be demonstrated in this tutorial.

## Prerequisites

You need access to a Postgresql database and sufficient privileges to create schemas and
roles. For more information see [prerequisites](prerequisites.md).

Your shell environment should have the appropriate Postgresql environment variables set. These are defined in
the [Postgresql documentation](https://www.postgresql.org/docs/current/libpq-envars.html). If you can run `psql`,
then `pgpkg` should work.

## Installing pgpkg

Install pgpkg:

    $ go install github.com/pgpkg/cmd/pgpkg

## Creating a database

Create a database for the tutorial:

    $ createdb hello
    $ export PGDATABASE=hello

## Writing your first function

Create a folder for your project:

    $ mkdir hello-pgpkg
    $ cd hello-pgpkg

Each pgpkg package requires a small configuration file.  Create one called `pgpkg.toml` file
in the folder project folder:

    Package = "github.com/example/hello-pgpkg"
    Schemas = [ "hello" ]

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

    $ pgpkg deploy

(if you want to see what `pgpkg` actually does, use `pgpkg --verbose .`)

If all goes well, you will now have a function defined in your database:

    $ psql
    psql> select hello.func();
    NOTICE:  Hello, world!
    func
    ------
    
    (1 row)

While it worked, it turns out there's a bug in our code. The function printed a message to the console
instead of returning the message as a value. Let's fix that!

With traditional migration tools, you would need to create a new version of the function
as an upgrade. With `pgpkg` you can simply edit the existing definition. So, edit func.sql:

    create or replace function hello.func() returns text language plpgsql as $$
      begin
        return 'Hello, world!';
      end;
    $$;

Apply the changes to the database:

    $ pgpkg deploy

And run it again:

    $ psql
    psql> select hello.func();
    func
    ---------------
    Hello, world!
    (1 row)

That's it! You've written your first pgpkg application - without writing a single migration
script.

In `pgpkg`, the function `hello.func` is called a _managed object_ (sometimes abbreviated
to _MOB_). Managed objects don't need migration scripts; you can treat them just like
you would any other code.

This is the main benefit of `pgpkg`: it makes working with functions, views and triggers
much easier.

## Creating a database table

Database tables are _migrated objects_, which means they still need to be created and updated
using traditional migration scripts. Let's create one!

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

    $ pgpkg deploy

Let's see if the table exists:

    $ psql
    psql> select * from hello.contact;
    name
    ------
    (0 rows)
    
We forgot to populate the table with some default values! So, let's add another migration script.
Call it contact@001.sql, because it's the first change to the contact table. Edit the
file `schema/contact@001.sql`:

    insert into hello.contact (name) values ('Postgresql Community');

Remember that pgpkg needs to know the order in which migrations will run, so you
need to add this new migration script to `schema/@migration.pgpkg`. It should now
look like this:

    contact.sql
    contact@001.sql

You can again apply the updated package to the database:

    $ pgpkg deploy

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
order, with `contact.sql` followed by `contact@001.sql`. This makes it much easier to
understand how objects have changed over many migrations - especially when migrations
occur over several years.

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

    $ pgpkg deploy

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
tests in the migration directory, though).

Try adding this script to `world_test.sql` in your project:

    create or replace function hello.world_test() returns void language plpgsql as $$
        begin
            if hello.world() <> 'Postgresql Community' then
                raise exception 'the world is not right';
            end if;
        end;
    $$;

As usual, apply the changes to the database:

    $ pgpkg deploy

The test will have been applied, but you won't see anything if it passes, because `pgpkg` doesn't log much, unless
something goes wrong.

To see if the tests are working, use `--show-tests`:

    $ pgpkg deploy --show-tests
    pgpkg: 2023/08/11 15:16:40   [pass] pgpkg.op_test()
    pgpkg: 2023/08/11 15:16:40   [pass] hello.world_test()

You can see that one test ran in your package (the other package is `pgpkg` itself).

You can add `raise notice` (and `raise warning`) commands to your tests to log information
to the console during the testing process. Edit `world_test.sql` to add a notice:

    create or replace function hello.world_test() returns void language plpgsql as $$
        begin
            raise notice 'Testing the world';
            if hello.world() <> 'Postgresql Community' then
                raise exception 'the world is not right';
            end if;
        end;
    $$;

> Using JSON, you can even use `raise notice` to print the rows of a table or query:
>
>     raise notice '%', (SELECT jsonb_pretty(jsonb_agg(t)) FROM mytable t);

Let's install the new script, which will run the test and display the notice to the console:

    $ pgpkg deploy         
    [notice]: Testing the world

Any `raise notice` commands from tests that run will be printed to the console.
(`raise warning` commands are printed to stderr). Messages raised in migration
scripts will also be displayed.

A successful test is one that finds a problem - so let's create a problem!

    $ psql
    psql> update hello.contact set name = 'World';
    UPDATE 1

Now, reinstall the package, which will re-run the tests:

    $ pgpkg deploy        
    [notice]: Testing the world
    ./world_test.sql:1: test failed: hello.world_test(): pq: the world is not right
           3:         raise notice 'Testing the world';
           4:         if hello.world() <> 'Postgresql Community' then
    -->    5:             raise exception 'the world is not right';
           6:         end if;
           7:     end;
    PL/pgSQL function world_test() line 5 at RAISE

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