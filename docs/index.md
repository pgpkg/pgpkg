# pgpkg - a package manager for pl/pgsql

![pgpkg logo](logo-small.png)

[Scroll to documentation](#documentation)

## What is pgpkg?

pgpkg is a small and fast command-line package manager for pl/pgsql, the built-in Postgres programming language.
It's a single binary, written in Go, and tries to make minimal assumptions about your database beyond what
pgpkg touches. It should also work with any Postgresql language.

pgpkg is designed to make writing Postgresql stored functions as easy as writing functions in a language like
Go, Java or Python. It eliminates the need to write migration scripts for functions, views and triggers. You can
simply edit these objects, and pgpkg will deal with upgrading them. If you like, you can mix SQL files with your
regular code, placing like-named files together in your directory tree.

In addition to managing functions, views and triggers, pgpkg also supports traditional
migration scripts for tables, UDTs and other long-lived database objects.

You can use pgpkg to manage the schema for a single application, or you can use
it to import and manage multiple schemas, just like you might include dependencies in
a Go, Java or Python program.

pgpkg installs itself into a database, and does not require any extensions - pgpkg is itself
a pgpkg package. pgpkg and dependencies can be incorporated directly into Go programs as a
library, making a Go database self-migrating, and the binary and packages can be included
as part of deployments for other languages such as Java or Python.

pgpkg was first released on 23rd Feb 2023, and should be considered early alpha quality.
It's unlikely to break any databases, but it probably has bugs, and definitely isn't finished.
See [status](#status) for more information.

## What is pl/pgsql?

[pl/pgsql](https://www.postgresql.org/docs/current/plpgsql.html) is the native programming language for Postgresql.
It's not modern or fancy, but nothing can beat it if you want to build a quick business app, or a prototype,
or even a sophisticated data management app. In fact, if your application is mostly database operations,
pl/pgsql is likely to significantly reduce the number of lines of code you need to write to perform the
same operation in a host language like Go, Java or Python. And it's likely to run significantly faster, too.

However, one thing that's been missing from this language is a package manager.

pgpkg is my attempt to fix that.

## Downloading pgpkg

At the moment, the easiest way to try pgpkg is to [install Go](https://go.dev/dl/) (1.20 or later) and run:

    go install github.com/pgpkg/cmd/pgpkg

This will install pgpkg in your GOBIN directory. If that's in your `$PATH` then you're set.

## What it does

pgpkg aims to be:

* a directory structure specification for distributing Postgresql code in pl/pgsql or SQL.
* a simple, atomic, fast, safe, and easy schema migration tool
* a tool to download and manage remotely hosted code (and dependencies), and
* a tool to install the packages into a database

pgpkg also lets you define and run tests, all of which are required to pass before
a schema migration will be committed. Tests are written as pl/pgsql functions, similar
in style to Go tests.

## Documentation

* [Features](pages/features.md)
* [Installing Packages](pages/installing.md)
* [pgpkg Package Specification](pages/spec.md)
* [Installation Phases](pages/phases.md)
* [Writing attractive plpgsql](pages/plpgsql.md)
* [Safety and Security](pages/safety.md)

## Example Package

A basic example package can be [found here](https://github.com/pgpkg/pgpkg-test).

I hope to provide more detailed examples in the future.

## Status

pgpkg is **early alpha**. I use it for my own work every day, but there is still much to be done.
Major features that need implementing include:

* plpgsql doesn't yet support subdirectories in api, schema or tests. It's on the TODO list.
* dependency management, download & install. While you can have multiple pgpkg packages in a single
  database, and pgpkg manages isolation (to a point), each dependency must currently be installed
  manually.
* security. pgpkg makes some effort to isolate packages through the judicious use of roles,
  but this support is incomplete and easy to defeat.
* schema relocation. in the event that two packages use the same schema, it would be good to be able
  to place the packages in a schema other than the one defined by the package maintainer.
  Some work has been done towards making this possible, but it is incomplete.
* File locations. One goal of pgpkg is to allow SQL procedures to appear in your source tree
  next to your Go or Java code. But that hasn't happened yet.

## Inspiration

The inspiration for pgpkg is, perhaps naturally, the [Go programming language](https://go.dev) in which pgpkg is
written. The ultimate goal is to contribute to the growth of the plpgsql user community, by providing tools that enable
the creation and sharing interesting Postgresql database functionality, as easily as they can share Go code.