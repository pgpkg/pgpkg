# Tutorial Prerequisites

To usepgpkg, you need to be able to access a Postgresql database. You don't need
superuser access, but you do need privileges including the ability to create
database schemas and roles.

> NOTE: `pgpkg` uses the same `PGDATABASE` and other environment variables as `psql`.
> It does not yet support command-line options to override the environment. You can also
> use a `DSN` environment variable to use a URL-style connection string.

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

