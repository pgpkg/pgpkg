# pgpkg features

pgpkg eliminates the need to write migration scripts for functions, views  and triggers.
It does this by treating these objects  - and updating them if they change - in the same way that a
function or method in a Go, Java or Python can be changed.

Rather than having to write a migration script for functions, you just change the source code, and
pgpkg does the rest. It automatically resolves function dependencies and keeps your schema clean.

This is a different approach from most migration tools which treat functions, views and
triggers in the same way they treat tables and user-defined types. But that's unwieldy.
It can be really hard to find the current function definition in a migration system,
and I believe that this leads to functions being seen as difficult to use, when in fact
they require a lot less code and are much more performant than similar functions developed
outside the database.

## Parser Based

pgpkg uses Postgres' own parser to parse the SQL scripts defined in a package. This means
that pgpkg is able to understand function, view and trigger declarations, and it uses this
understanding to manage these objects.

In the future, the hope is to continue to extend the use of the parser to enable greater safety,
as well as the ability to implement important features such as schema rewriting.

## Schemas and Roles

pgpkg installs packages into individual schemas. Each schema is also assigned a unique database role,
which is intended to increase the isolation of packages from one another. Package roles are not
granted access to other schemas unless specifically requested in the package definition.

> Package isolation is a work in progress. Although we create package roles that are unprivileged,
> at this time you should assume that privilege escalation is possible during migrations.

## Notice logging

During installation and testing, pgpkg intercepts `raise notice` messages and prints them to the
console. This makes the command super useful for debugging.

## Contextual errors

When something goes wrong, pgpkg prints nice contextual error messages, such as this:

    tests/failing-tests/tests/test2.sql:17: test failed: failing_tests.test_5(): pq: test 5
           1:
           2: begin
    -->    3:     raise exception 'test 5';
           4: end;
           5:
    PL/pgSQL function test_5() line 3 at RAISE

Contextual errors are a work in progress!

## Atomic Upgrades

plpkg peforms **all** operations for an upgrade in a single transaction. An upgrade either completely
succeeds, or completely fails.

pgpkg uses a few tricks to acheive this. For example, functions, views and triggers may create dependencies on one
another. pgpkg will attempt to repeatedly install objects until all dependencies are met (or until progress stops).
To do this, we do a lot of work inside savepoints. But you don't have to care about that.

## Non destructive tests

After an upgrade has otherwise succeeded, plpgk will attempt to run any tests defined in the `tests`
directory. Tests are pl/pgsql functions whose name starts with `test_`. Tests are run in random order.

Tests are run inside a savepoint, and rolled back at completion. A test can fail simply by calling
`raise exception`. Future versions of pgpkg will probably include some simple testing tools.

## Rollback scripts

pgpkg does not support rollback scripts, and it's unlikely that it ever will. If a migration fails then
the transaction rolls back, no damage is done, and you can go back to using your old code. If a transaction
succeeds and then needs to be rolled back, you can just write another migration.

We do not support automatic recovery from migration failures because doing so adds complexity, and complexity
is likely to make upgrades less reliable.