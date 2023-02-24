# Writing Attractive pl/pgsql

pl/pgsql is easy to use, concise, and fast. For database operations, it is far more efficient,
and requires far less code, than performing the equivalent operations in a host language such
as Go, Java or Python.

Unfortunately, Most examples of pl/pgsql are written
[like this](https://www.postgresql.org/docs/15/plpgsql-statements.html#PLPGSQL-STATEMENTS-EXECUTING-DYN):

    CREATE OR REPLACE FUNCTION EMPLOY() RETURNS VOID LANGUAGE PLPGSQL AS $$
        BEGIN
            SELECT * INTO STRICT myrec FROM emp WHERE empname = myname;
            EXCEPTION
                WHEN NO_DATA_FOUND THEN
                    RAISE EXCEPTION 'employee % not found', myname;
                WHEN TOO_MANY_ROWS THEN
                    RAISE EXCEPTION 'employee % not unique', myname;
        END;
    $$;

This is a little guide to enjoying the experience of writing pl/pgsql.

## Use Lowercase

I've found that it's much easier to enjoy using pl/pgsql when we stop shouting:

    create or replace function employ() returns void language plpgsql as $$
        begin
            select * into strict myrec from emp where empname = myname;
            exception
                when no_data_found then
                    raise exception 'employee % not found', myname;
                when too_many_rows then
                    raise exception 'employee % not unique', myname;
        end;
    $$;

## Debugging

You can use `RAISE EXCEPTION` - I mean,

    raise exception '...';

for debugging. When you use this in a pgpkg test, the message is logged to the console when pgpkg runs.

You can include just about anything in a `raise exception` using the interpolation character (`%`):

    create or replace function example() returns void language plpgsql as $$
        declare
            a integer = 10;
            b decimal = 20;
            t text = 'hello, world';

        begin
            raise notice '%: a=% and b=%', t, a, b;
        end;
    $$;

## Use The Force

Postgresql has a lot of features that are rarely used. Among my favourites:

### User defined types

Postgres has a rich language for creating [user-defined types](https://www.postgresql.org/docs/15/sql-createtype.html).
Here are some examples.

You can define your own types, based on some other type; `domains` also allow you to put
constraints on a type. The type you constrain can also be a composite type. Example:

    create domain example.primary_key as uuid;

You can define enums; these require less storage space than strings and, in my opinion,
are preferable to arbitrary text values:

    create type example.transaction_e as enum ('invoice', 'receipt', 'debit', 'credit');

You can create composite types, known in Go and C as `structs`:

    create type example.report_t as (
        report report_k,
        currency currency_t,
        amount decimal
    );

A table is basically a composite type with a backing store, so almost anything you can do with a
table, you can do with a composite type. For example, you can pass them around to functions:

    create or replace function example(_r example.report_t) returns example.report_t ...

Note that in `pgpkg`, you must define these types as part of the `schema` definition, using migration
scripts.

### Arrays

Postgresql has great support for array values, which can make accessing certain kinds of
data much more efficient. For example:

    create table example.parent (
        child_names text[];
    );

    create table example.transactions (
        transaction_types example.transaction_e[];
    );


## Naming conventions and Style Guide

Having a reasonable naming convention is important with plpgsql because the way the language works makes it
unfortunately easy to create naming conflicts. For example, if you have a function parameter and a table
with the same name, you will run into difficulties:

    create table report (
        r integer
    );

    create or replace function example(report integer) returns void language plpgsql as $$
        declare
            id integer;
        begin
            select r into id from report;
        end;
    $$;

This function won't work because the `report` argument shadows the `report` table. If there's an `id` table,
then that will probably cause you trouble as well.

To resolve this problem, simply make sure you use a regular naming scheme.

### Declare arguments and local variables using an underscore prefix:

    create or replace function example(_report_k integer) returns void language plpgsql as $$
        declare
            _id integer;
        begin
          select r into _id from report where report = _report_k;
        end;
    $$;

### Use a naming convention for data types.

Variable names can also become ambiguous; for example, your `account` table has a primary key, and you will
often pass the primary key around. Since `account` is the name of a table (as well as the name of the type
associatd with that table), is a variable called `_account` the value of a row, or the value of its primary key?

In fact, you'll often find that you pass in the key to a table, and then need to look it up. So, to make code
unambiguous, I use the following scheme:

* Suffix the name of user-defined composite types with `_t`. For example: `report_t`.  
* Suffix the name of types used in keys with `_k`. Example: `report_k`. 
* Suffix the name of enums with `_e`. Example: `transaction_e`.

Data types that don't have a suffix - such as `report` - are assumed to be table names.

    create table report (
        report   report_k,
        currency currency_t,
        amount   decimal
    );

    create or replace function func(_report_k report_k) returns void language plpgsql as $$
        declare
            _report report;

        begin
            select * into _report from report where report = _report_k;
        end;
    $$;

*Oh, yeah.*

## Schemas

`pgpkg` requires that you use schemas for your packages, but it's a good idea for your code anyway:

    create schema example;

    create domain example.report_k as uuid;
    create domain example.currency_t as text;

    create table example.report (
        report   example.report_k,
        currency example.currency_t,
        amount   decimal
    );

    create or replace function example.func(_report_k example.report_k) returns void language plpgsql as $$
        declare
            _report example.report;

        begin
            select * into _report from example.report where report = _report_k;
        end;
    $$;
