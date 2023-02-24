# pgpkg - a package manager for Postgresql

pgpkg is a small and fast package manager for pl/pgsql, the built-in Postgres programming language.
It's a single binary (written in Go), and tries to make minimal assumptions about your database beyond what
pgpkg itself touches.

## Why pgpkg

Unlike regular migration tools, pgpkg is designed to make writing Postgresql stored functions as easy as writing
functions in a language like Go, Java or Python.

**pgpkg eliminates the need to write migration scripts for functions,
views and triggers**. You can simply edit these objects in your regular text editor, and pgpkg will deal
with upgrading them.

pgpkg supports traditional, incremental migration scripts for tables, UDTs and other objects.

You can use pgpkg to manage the schema for a single application, or you can use
it to import and manage multiple schemas (coming soon), just like you would include dependencies in
a Go, Java or Python program.

pgpkg installs itself into a database, and does not require any extensions - pgpkg is itself
a pgpkg package! pgpkg and dependencies can be incorporated directly into Go programs as a
library, making a Go database self-migrating. The binary and packages can be included
as part of deployments for other languages such as Java or Python.

## Status

pgpkg is early alpha. I use it all the time, but YMMV.

## Documentation

Initial set of documentation is [here](docs/index.md).

## License

pgpkg is [licensed](LICENSE) under the same terms as Postgresql itself.

## Contributing

Contributors welcome.

