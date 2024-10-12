# pgpkg - a package manager for pl/pgsql

![pgpkg logo](logo-small.png)

## What is pgpkg?

pgpkg is a small, fast and safe Postgresql database migration command-line tool (and Go library) focussed on:

* keeping logic in one place, by enabling SQL functions and native code to be stored in the same directory;
* reusing logic, by enabling creation of SQL libraries for reuse in different systems.
* simplicity and ease of use.

pgpkg makes writing Postgresql stored functions as easy as writing functions in any other language, such as Go, Java or Python.

pgpkg enables your SQL code to live side-by-side with your regular code, and lets you use the exact same
workflows. You can edit your SQL functions in the same IDE, commit them to the same Git repository,
review them with PRs alongside other changes, and deploy them seamlessly to production.

## Documentation

* [Pgpkg Manual](pages/manual.md)
* [Pgpkg Tutorial](pages/tutorial/tutorial.md)
* [Why stored procedures?](#why-stored-procedures)
* [What is pl/pgsql?](#what-is-plpgsql)
* [Downloading pgpkg](#downloading-pgpkg)
* [pgpkg Status](#status)
* [pgpkg Features](pages/features.md)
* [Installing Packages](pages/installing.md)
* [pgpkg Package Specification](pages/spec.md)
* [pgpkg Installation Phases](pages/phases.md)
* [Writing attractive plpgsql](pages/plpgsql.md)
* [Safety and Security](pages/safety.md)

## Why stored procedures?

Postgresql plpgsql stored procedure are easy to write, require less code, and execute far
more quickly than the equivalent code running in a remote process. Until now, however, they were
a pain to manage.

## What is pl/pgsql?

[pl/pgsql](https://www.postgresql.org/docs/current/plpgsql.html) is the native programming language for Postgresql.
It's not modern or fancy, but nothing can beat it if you want to build a quick business app, or a prototype,
or even a sophisticated data management app. In fact, if your application is mostly database operations,
pl/pgsql is likely to significantly reduce the number of lines of code you need to write to perform the
same operation in a host language like Go, Java or Python. And it's likely to run significantly faster, too.

## Downloading pgpkg

At the moment, the easiest way to try pgpkg is to [install Go](https://go.dev/dl/) (1.23 or later) and run:

    go install github.com/pgpkg/pgpkg/cmd/pgpkg@latest

This will install pgpkg in your GOBIN directory. If that's in your `$PATH` then you're set. See the
[tutorial](pages/tutorial/tutorial.md) for more information.

## Status

pgpkg is **early alpha**. I use it for my own work every day, but there is still much to be done.
Major features that need implementing include:

* dependency management, download & install. While you can have multiple pgpkg packages in a single
  database, and pgpkg manages isolation (to a point), each dependency must currently be installed
  manually.
* security. pgpkg makes some effort to isolate packages through the judicious use of roles,
  but this support is incomplete and easy to defeat.
* schema relocation. in the event that two packages use the same schema, it would be good to be able
  to place the packages in a schema other than the one defined by the package maintainer.
  Some work has been done towards making this possible, but it is incomplete.

## Inspiration

The inspiration for pgpkg is, perhaps naturally, the [Go programming language](https://go.dev) in which pgpkg is
written.

The ultimate goal of pgpkg to contribute to the growth of the plpgsql user community and language, by providing
tools that enable the creation and sharing interesting Postgresql database functionality, as easily as they can
share Go, Java or Python code.