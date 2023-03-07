# Installing pgpkg packages

The intent of pgpkg is to allow the download and installation of plpgsql code using services like GitHub.
However, the ability to download packages isn't complete yet.

In the meantime, you can download a package and install it manually.

## Prerequisites

> pgpkg is in early alpha state, so we don't have any binaries available yet.

1. If you need to, download and [install Go](https://go.dev/dl/) (1.20 or later).
2. run: `go install github.com/pgpkg/cmd/pgpkg`

## Database Access

The `pgpkg` command uses the standard `PGXX` variables to connect to a Postgres server. If you can use `psql`,
`pgpkg` should work for you.

> `pgpkg` needs to run as a privileged database user. It drops privileges as it installs packages. For more
information, [see the safety page](safety.md).

## Download a pgpkg package

A `pgpkg` package is any directory which contains a `pgpkg.toml` file. Here's one we prepared earlier:

    git clone https://github.com/pgpkg/pgpkg-test
    cd pgpkg-test
    
## Install it.

Installing a package is super easy:

    pgpkg .

If the tests print anything with `raise notice`, you will see them on the terminal:

    [notice]: test 5
    [notice]: test 1
    [notice]: test 2
    [notice]: api 1
    [notice]: test 3
    [notice]: test 4

Unless you see an error message, the installation worked.

You can also use -verbose if you want to look under the covers a bit (not quite everything is logged yet):

    pgpkg -verbose .

which will give you detailed information about the SQL it executes:

    select count(*) from pg_roles where rolname=$1 [pgpkg]
    create schema if not exists "pgpkg" authorization "pgpkg"
    [notice]: schema "pgpkg" already exists, skipping
    note: github.com/bookwork/pgpkg: no MOB defined
    set role "pgpkg"
    reset role
    insert into pgpkg.pkg (pkg, schema_name, uses) values ($1, $2, $3) on conflict (pkg) do update set schema_name=excluded.schema_name, uses=excluded.uses [github.com/bookwork/pgpkg pgpkg 0x1400000e0c0]
    set role "pgpkg"
    reset role
    select count(*) from pg_roles where rolname=$1 [pgpkg_test]
    create role "pgpkg_test"
    create schema if not exists "pgpkg_test" authorization "pgpkg_test"
    parsing MOB api_1.sql
    set role "pgpkg_test"
    reset role
    set role "pgpkg_test"
    reset role
    insert into pgpkg.pkg (pkg, schema_name, uses) values ($1, $2, $3) on conflict (pkg) do update set schema_name=excluded.schema_name, uses=excluded.uses [github.com/pgpkg/pgpkg-test pgpkg_test 0x140004444b0]
    set role "pgpkg_test"
    parsing tests tests/test1.sql
    parsing tests tests/test2.sql
    [notice]: test 3
    [PASS] pgpkg_test.test_3()
    [notice]: test 4
    [PASS] pgpkg_test.test_4()
    [notice]: test 5
    [PASS] pgpkg_test.test_5()
    [notice]: test 1
    [PASS] pgpkg_test.test_1()
    [notice]: test 2
    [PASS] pgpkg_test.test_2()
    [notice]: api 1
    [PASS] pgpkg_test.test_6()
    reset role
    installed 1 function(s), 0 view(s) and 0 trigger(s). 1 migration(s) needed. 6 test(s) run
