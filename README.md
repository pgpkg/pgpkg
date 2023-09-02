# pgpkg - a schema and package manager for Postgresql

![pgpkg logo](docs/logo-small.png)

pgpkg is a small and fast command-line tool (and Go library) that lets pl/pgSQL functions live
side-by-side with regular code, allowing you use the exact same IDE (or non-IDE) development workflows for SQL and non-SQL
code.

You can edit your SQL functions in the same IDE, commit them to the same Git repository,
review them with PRs alongside native code changes, and deploy them seamlessly to production.

pgpkg also lets you package up your SQL code and incorporate it as a dependency into other projects,
similar to Node packages or Go modules, but note that dependency support is still early days.

## Documentation

The best place to start is [the pgpkg man page](docs/pages/manual.md).

## Tutorial

The [tutorial for using pgpkg](docs/pages/tutorial/tutorial.md) contains a worked example for
writing functions, unit tests and migration scripts. If you're familiar with Postgresql,
it will only take a few minutes to work through.

## Status

pgpkg is late alpha. I apologise in advance if documentation or examples are out of date.

I use pgpkg pretty much on a daily basis. It works really well for me, and I'm working to remove
the rough edges.

I work on making it easier to use when I have time. The [TODO](TODO.md) list contains issues that I
expect to fix over time.

PRs and issues are welcome.

## Documentation

Initial set of documentation is [here](docs/index.md). I have been focusing on writing the manual and tutorial,
so other documents may currently be out of date.

## License

pgpkg is [licensed](LICENSE.md) under the same terms as Postgresql itself.

## Contributing

Contributors welcome. Contact me at [mark@commandquery.com](mailto:mark@commandquery.com)

