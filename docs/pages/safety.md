## Safety and Security

> **BE WARNED** that these are early days for pgpkg. I have made an attempt to
> provide a framework to enable security via Postgresql primitives (such as
> the use of roles and schemas), but at this time you should consider any package you use to
> potentially have the ability to read and write to your database.
> 
> pgpkg is great for reuse of internally developed code, but **an adversary would certainly
> be able to defeat the security measures currently in use at this time**.

## Introduction

Creating a safe and secure environment inside an existing database in order to allow the download
of software written on the internet is challenging. It looks like it may be possible, but will
take some effort.

pgpkg has implemented a number of features which are intended to provide a reasonable foundation for
improving security in future versions. But additional measures will be necessary before pgpkg could be
considered mature enough to download code from the internet, and execute it (!).

## Known Issues

pgpkg is a work in progress. Although we parse all SQL using Postgres' own parser, we don't
sanitize it much yet.

Some known issues include:

Migration scripts can run *any* SQL - in particular:
* `commit work` / `rollback work`
* `reset role` and possibly `set role`
* `execute`

Managed objects and tests can run:
  * `reset role` and the equivalent `authorization` commands
  * `execute`

Some DDL commands may not yet sanitise their inputs (TODO).

## Roles

When a request is made to install a package into a particular schema, `pgpkg` creates both a schema
and an associated role for the package. The role name is the schema name, prefixed with `$`.
For example, schema `finance` is associated with the role `$finance`. This is intended to
reduce the likelihood of collisions with existing or future role names.

The package's schema is owned by its role, which means that the role can be used to create and access
objects within the schema, but not to access other schemas.

## Privilege Dropping

During installation of a package, `pgpkg` drops privileges when it runs code from the package.
Unfortunately, at the moment a package can trivially restore privileges using `reset role`.
This means that packages can easily escalate their privileges to the level of the user
who invoked `pgpkg`.

It is likely that this problem can be resolved, but we're not there yet.

## Schema search path

Every function created by `pgpkg` has the `search_path` set to "pgpkg", "pg_temp", "public".
The search path is also set during migrations and tests.

Note that it's easy to change the `search_path`, and even without changing it, you can simply write code
that refers to any other schema. However, this approach is thwarted during upgrades because the schema role
doesn't have access to other schemas.

## Security Definer

To ensure that functions can't access objects outside the package (schema) in which they are
defined, functions are owned by the schema and are declared with the  `security definer`
option.

This behaviour will be extended to views in the future.

## Automatic `grants`

When a package refers to another package (using the `Uses` list in `pgpkg.toml`), `pgpkg` will grant access
to the schema and its objects to the enclosing package.

For example, if package `parent` specifies

    Uses: [ "child" ]

then the following will happen (summarised):

    create role parent;
    create schema parent authorization parent;
    create role child;
    create schema child authorization child;
    ...
    grant usage on schema "child" to "parent"
    grant execute on all functions in schema "child" to "parent"
    grant select, update, insert, references on all tables in schema "child" to "parent"
    grant usage on all sequences in schema "child" to "parent"

This effectively gives access to the child schema without allowing the parent schema to modify it.

> These grants are deliberately broad, but in later versions of pgpkg we may wish to hide data
> inside schemas. That's a discussion for another day.