## Deploying with Go

One of pgpkg's features is that it plays nicely with other languages. It plays especially
nicely with Go, but you can use the features of pgpkg with any language.

In this example we're going to integrate our application with a Go program.

First, you need to create a module and add pgpkg as a dependency:

    $ go mod init github.com/pgpkg/example
    go: creating new go.mod: module github.com/pgpkg/example
    go: to add module requirements and sums:
    go mod tidy
    $ go get github.com/pgpkg/pgpkg

Next, we need to set up main.go to load the schema and open the database:

    package main

    import (
        "embed"
        "fmt"
        "github.com/pgpkg/pgpkg"
    )
    
    //go:embed pgpkg.toml *.sql schema
    var hello embed.FS

    func main() {
        pgpkg.ParseArgs()
        var p pgpkg.Project
        p.AddFS(hello)
    
        db, err := p.Open()
        if err != nil {
            pgpkg.Exit(err)
        }
        defer db.Close()
    
        var world string
        if err = db.QueryRow("select hello.world()").Scan(&world); err != nil {
            pgpkg.Exit(err)
        }
    
        fmt.Println("And the world is:", world)
    }

Running this program will do the following:

* parse command line options like "--verbose"
* run any necessary migration
* install the SQL function
* run the tests, and
* return a database handle.

So let's build and run it:

    $ go build
    $ ./example
    [notice]: Testing the world
    And the world is: Postgresql Community

Note how the test has printed a message - `raise notice` commands are automatically
logged from pgpkg. We can fix this by updating the test. Edit world_test.sql to remove
the notice:

    create or replace function hello.world_test() returns void language plpgsql as $$
        begin
            if hello.world() <> 'Postgresql Community' then
                raise exception 'the world is not right';
            end if;
        end;
    $$;

Then build and rerun your code:

    $ go build
    $ ./example
    And the world is: Postgresql Community

> Note that you don't need to run `pgpkg` when you've embedded your package into your Go
> program. The Go program migrates the database for you, when you call p.Open().

`pgpkg.ParseArgs` adds support for a standard set of migration options to your Go program.
These include:

    --summary
    --verbose
    --dry-run

For example, we can run this:

    $ ./example --summary
    github.com/bookwork/pgpkg: installed 0 function(s), 0 view(s) and 0 trigger(s). 0 migration(s) needed. 0 test(s) run
    github.com/example/hello-pgpkg: installed 2 function(s), 0 view(s) and 0 trigger(s). 0 migration(s) needed. 1 test(s) run
    And the world is: Postgresql Community

Finally, note how your SQL and Go code can live in the same directory:

    .
    ├── example
    ├── func.sql
    ├── go.mod
    ├── go.sum
    ├── main.go
    ├── pgpkg.toml
    ├── schema
    │   ├── @migration.pgpkg
    │   ├── contact.sql
    │   └── contact@001.sql
    ├── world.sql
    └── world_test.sql

This makes switching between Go and plpgsql code in your IDE completely seamless.

