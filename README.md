# pgpkg - a package manager for Postgresql

![pgpkg logo](docs/logo-small.png)

pgpkg is a small and fast command-line tool (and Go library) which is designed to make writing Postgresql
stored functions as easy as writing functions in any other language, such as Go, Java or Python.

At it's core, pgpkg is a database migration tool that reduces the hassle of writing and shipping stored
functions and other static objects in Postgresql.

pgpkg lets your SQL functions live side-by-side with your regular code, and lets you use the exact same
workflows. You can edit your SQL functions in the same IDE, commit them to the same Git repository,
review them with PRs alongside other changes, and deploy them seamlessly to production.

## Tutorial

The [tutorial for using pgpkg](docs/pages/tutorial/tutorial.md) contains a worked example for
writing functions, unit tests and migration scripts. You can see the example
code created in the tutorial [here](https://github.com/pgpkg/pgpkg/tree/main/tests/good/example).

## Status

pgpkg is early alpha. I use it all the time, but YMMV.

## Documentation

Initial set of documentation is [here](docs/index.md).

## License

pgpkg is [licensed](LICENSE.md) under the same terms as Postgresql itself.

## Contributing

Contributors welcome.

