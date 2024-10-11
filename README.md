# pgpkg - simplifies Postgresql stored procedure development.

![pgpkg logo](docs/logo-small.png)

pgpkg is a small and fast command-line tool (and Go library) that lets your pl/pgSQL functions live
side-by-side with regular code, allowing you to use the exact same workflows, source code control,
IDE (or non-IDE) and other development workflows for both SQL and non-SQL code. It automatically
deploys your functions without the need to maintain migration scripts.

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

## Release Notes

* 2024-10-11 Added `Migrations` clause to `pgpkg.toml`. `@migration.pgpkg` is now deprecated.
  [See the faq](docs/pages/faq.md#what-happened-to-migrationpgpkg) for more details.
* 2024-10-11 --force-role option allows the default package role name to be overridden.
* 2024-10-11 removed character set restriction from schema names in TOML files.
