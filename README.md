# pgpkg - a schema and package manager for Postgresql

![pgpkg logo](docs/logo-small.png)

pgpkg is a small and fast command-line tool (and Go library) that lets pl/pgSQL functions live
side-by-side with regular code, allowing you use the exact same workflows for SQL and non-SQL
code.

You can edit your SQL functions in the same IDE, commit them to the same Git repository,
review them with PRs alongside other changes, and deploy them seamlessly to production.

pgpkg also lets you package up your SQL code and incorporate it as a dependency into other projects.

## Documentation

The best place to start is [the pgpkg man page](docs/pages/manual.md).

## Tutorial

The [tutorial for using pgpkg](docs/pages/tutorial/tutorial.md) contains a worked example for
writing functions, unit tests and migration scripts. You can see the example
code created in the tutorial [here](https://github.com/pgpkg/pgpkg/tree/main/tests/good/example).

## Status

pgpkg is early alpha. I apologise in advance if documentation or examples are out of date.

I use pgpkg pretty much on a daily basis. It works really well for me, but it does have rough edges.
I work on making it easier to use when I have time. The [TODO](TODO.md) list contains issues that I
expect to fix over time.

PRs and issues are welcome.

## Documentation

Initial set of documentation is [here](docs/index.md).

## License

pgpkg is [licensed](LICENSE.md) under the same terms as Postgresql itself.

## Contributing

Contributors welcome.

