# pgpkg FAQ

## General Questions

### What is pgpkg?

pgpkg is a command-line tool (and Go library) which makes it easy to write and deploy pl/pgsql database functions and
migration scripts. It also allows you to package and share SQL code - such as function libraries
and even database tables - between your projects.

### Why is pgpkg different from other database migration tools?

pgpkg is different in the following ways:

* it lets you declare SQL functions within your IDE, right next to your regular code.
* it's intended exclusively for Postgresql (it uses Postgresql's SQL parser); it won't work with other SQL dialects.
* It *doesn't require migration scripts for functions, views and triggers*.
* it tries to be as simple and non-intrusive as possible, with no special filenames or funky script delimiters.
* it operates atomically - migrations either fully succeed, or fail.
* it supports the writing of both constructive and destructive tests.
* it requires the use of database schemas.

### My friends all tell me that database functions are stupid. Why should I care about them?

Friends are to be cherished, so you can still care about them despite their silly opinions.

Much of the FUD around database functions is because they are perceived to be difficult to maintain. With pgpkg,
I set out to make it as easy to work with database functions as it is to work with functions in any other language.

Once you eliminate the maintenance problems, plpgsql functions have many benefits. Executing code within
the database is much faster than moving data to your code, and plpgsql functions are far less verbose than similar
code written in another language, because plpgsql can take advantage of existing declarations, such as table
definitions, without the need for separate DTOs, DAOs, entity objects, and all that boilerplate.

### Am I required to use database schemas with pgpkg?

Yes.

### Do my SQL scripts need to have specific file names or patterns?

Not really.

Scripts containing SQL need names that end in `.sql`, and all pgpkg scripts must appear in or below a directory
that contains a `pgpkg.toml` file.

That said, typically `pgpkg.toml` would be declared in the root of your code tree (for example, alongside `.git`
or `package.json` or `go.mod`), which means that, in practice, your SQL functions can appear anywhere in your
code tree.

This flexibility lets you declare SQL functions in files that sit next to your regular code, which makes
flicking between them in your IDE super easy. For example, you can mix Go (or whatever language) and SQL files
in the same folder, such as:

    delete.go
    delete.sql

The only small exception to this rule is migrations. Migrations are scripts listed in the `Migrations` clause of
`pgpkg.toml`. The file *names* (without the path) of these scripts are used to determine if they have been
run previously, or not. This means that the name of these scripts cannot be changed once they've been deployed
in the field. You can, however, safely move the scripts to different folders (provided you update the `Migrations`
clause to point to the new location of the file).

There are no other unusual restrictions on the names of files.

### Do I need to be a Go developer to use pgpkg?

No! pgpkg is intended to be useful as a generic command line tool; it doesn't matter what language the rest
of your application is written in.

However, pgpkg can be used as a library within Go programs, meaning that you don't need to ship the
pgpkg utility separately, if you're using Go. You can also embed your SQL scripts in your Go code.

### Is pgpkg an ORM?

No. If anything, pgpkg is a reaction to ORMs. Put simply: if you write your database logic using SQL functions
then the whole problem that ORMs are supposed to solve, goes away.

### Can pgpkg generate stubs?

Stubs would be great because they would create native language bindings to database functions.

pgpkg doesn't generate stubs at the moment, but I hope to add this functionality one day. This would allow you to
call SQL functions as if they were written in Go (or Java, or Python, or ...).

### Aren't SQL functions hard to maintain?

Most migration tools that I've seen treat SQL functions the same as any other database object, which means
that a small change to a function requires a migration script.

The purpose of pgpkg is to treat functions (along with views and triggers) as if they were regular code.
This means that writing an SQL function is as easy as writing a function in any other language.

### How do I call the SQL functions?

Use your host language's SQL library to connect to the database and call the functions. For example,
in Go you might use the `database/sql` package, and call functions from `db.Exec()` or `tx.Exec()`.
In Java, you might use JDBC.

### Isn't plpgsql slow?

plpgsql is not designed to be a fast, general-purpose language, and CPU-only tasks can certainly be
slower than other languages.

However, for logic that primarily transforms a database, plpgsql is extremely fast. I've seen 10x to 100x better
performance using plpgsql instead of accessing a database over a network using another language. In addition,
plpgsql can perform some types of SQL optimisation which are just not possible outside the database.

Put another way: compared to the typical ORM sequence of assembling and sending a query over the network,
having it parsed, executing it, serialising the results, shipping them back over a network to your code,
decoding them into objects (or whatever), determining what changes you want to make, and then repeating the whole 
process for each object that you're changing... plpgsql is much faster.

Using plpgsql is also simpler in many ways, because it reduces the amount of boilerplate (DTOs, DAOs, entity objects),
and largely obviates the need for caching in your code. Instead, plgpsql can use existing table definitions as data
types, and uses the database cache, which is super efficient. In my experience, code written in plgpsql is not
just faster, but there's also a lot less of it.

### DO I HAVE TO USE CAPS IN MY CODE?

There's no need to shout! Using all-caps is not necessary, and I don't recommend using
capitals in any part of SQL scripts.

So to be clear, I do this:

    create or replace function f() returns integer language sql as $$ 
      select count(*) from schema.table where column = value;
    $$;

and I do not ever do this:

    CREATE OR REPLACE FUNCTION F() RETURNS INTEGER LANGUAGE SQL AS $$ 
      SELECT * FROM schema.table WHERE column = value;
    $$;

because 1970 called, and wants its teletypes returned. And then 1980 called and asked for
its low-res character generator ROMs back. And so on. You get the idea. We like
fast, efficient, readable code, and we are not barbarians.

### Does pgpkg support rollback scripts?

No. Rollbacks dramatically increase the complexity of a process which should be kept as simple as possible.

I've managed thousands of database migrations across hundreds of installations, and I can count the number of
migration problems I've had on one hand. (It takes no hands to count the number of total migration failures).
If migrations are fragile, you're doing something wrong... at least with Postgresql.

In fact, I think rollback migration scripts are a uniquely stupid invention. They were *arguably* necessary for
databases which could not do transactional DDL, but Postgres does not have this problem. Even without transactional
DDL, I'm not convinced they're a good idea. Trying to automate away a failure of a migration script is
like blowing out a fire with TNT. Sure, it's *possible*, but there's a good chance it will make things worse...
maybe a lot worse.

The really big problem with rollback scripts is that, by definition, they are only ever going to run when something
unexpected happens. You would not have run a migration in the first place if you expected it to fail. So how can
a rollback script, written at the same time as the script that's broken, know what to do?

So instead of rollback scripts, pgpkg provides you with unit testing facilities which are executed after
every migration. If the unit tests fail, the migration is cancelled and the database is left unchanged.

Rather than focusing on automation, make sure that any migration changes you make are non-destructive.
Use referential integrity. Dropping a column or table? Doing a large update? Copy the table to an archive
table before you drop it, so you can recover it manually if you need to. Remove old archive tables in
future migrations. Write SQL unit tests.

And make frequent backups. You know, just in case. 

## pgpkg changes

### What happened to @migration.pgpkg?

> `@migration.pgpkg` is now considered deprecated.

Early versions of `pgpkg` used a special file, `@migration.pgpkg`, to list migration scripts.
The directory for this file was treated as special.

`@migration.pgpkg` has been replaced with the `Migrations` clause in `pgpkg.toml`. 

To convert from `@migrations.pgpkg` in a directory called `schema` that contains the following:

    table1.sql
    table2.sql
    table3.sql

simply add a `Migrations` clause to your `pgpkg.toml` file:

    Migrations = [
        "schema/table1.sql",
        "schema/table2.sql",
        "schema/table3.sql",
    ]

and remove the `@migration.pgpkg` file.

`pgpkg` uses only the filename part of the path to determine if a migration script has been run already.
This means that after upgrading from `@migration.pgpkg`, you can move your migration scripts anywhere in your
directory tree.