## Deploying with Go

One of pgpkg's features is that it plays nicely with other languages. It plays especially
nicely with Go, but you can use most features of pgpkg with any language by calling the CLI.

## Integrating with Go

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
    │   ├── contact.sql
    │   └── contact@001.sql
    ├── world.sql
    └── world_test.sql

This makes switching between Go and plpgsql code in your IDE completely seamless.

## Unit Testing with pgpkg and Go

One of the more painful aspects of unit testing in database applications is setting up the data
for testing. `pgpkg`'s testing framework can help with this by helping with the creation of temporary
test databases, and providing a mechanism for calling test functions within a unit test.

### Test Setup

Setting up a database for unit testing is pretty easy. Here's an example script that creates a temporary
database, migrates it to the current version, creates some test data, runs some Go tests, and drops the database
when it's done. A script like this will typically complete in a few hundred ms:

Let's call this `app_test.go`:

```
//go:embed pgpkg.toml schema *.sql
var pgpkgSchema embed.FS

func TestSomeLogic(t *testing.T) {

    // Don't delete test scripts (see below).
    pgpkg.Options.KeepTestScripts = true

    var err error

    // Create a temporary database for this test. The DSN "" means we
    // will connect to the database in PGDATABASE and friends, to create
    // the temporary database that will be used in the test.
    tempDb, err := pgpkg.CreateTempDB("")
    if err != nil {
    	t.Fatalf("unable to create temp db: %v", err)
    }
    
    // Tell pgpkg the name of the database we're going to use.
    var dsn = "dbname=" + tempDb
    
    // Automatically drop the database when we're done.
    defer func() {
    	PGXPool.Close()
    	err := pgpkg.DropTempDB("", tempDb)
    	if err != nil {
    		t.Fatalf("unable to drop temporary database %s: %v", dsn, err)
    	}
    }()
    
    // Initialise pgpkg
    if err := pgpkg.ParseArgs("pgpkg"); err != nil {
    	pgpkg.Exit(err)
    }
    
    project := pgpkg.NewProject()
    if _, err := project.AddEmbeddedFS(pgpkgSchema, "."); err != nil {
    	pgpkg.Exit(err)
    }
    
    // Migrate the database to the current version.
    if err := project.Migrate(Config.DatabaseDSN); err != nil {
    	pgpkg.Exit(err)
    }
    
    // Set up your data.
    _, err = sql.Exec("select app.init_data()")
    
    // Run your tests
    // ...
}
```

The test is run using the normal `go test` commands.

## Using KeepTestScripts

> WARNING: only use KeepTestScripts with temporary, disposable databases. Migrations performed using
> KeepTestScripts **cannot be upgraded** after being created.

When running tests, `pgpkg` creates all the objects listed in all the `_test.sql` files that it finds. It then runs all
functions with names ending in `_test` as unit tests. Function names not ending in `_test`, and all views and triggers,
are available to the unit tests as library objects, for example to set up testing data or perform checks.

In a typical migration, all the unit test and utility functions are automatically deleted after running successfully.
However, by setting the `KeepTestScripts` option, these objects are retained after the migration is complete, and
can therefore be used in your Go unit tests.

> The --keep-test-scripts option can be used with the pgpkg CLI to retain test functions for non-go
> language tests.

By putting our data loading functions into `_test.sql` scripts and setting `KeepTestScripts` on a temporary
database for our Go unit test, we can create complex data loading functions in SQL as part of our Go unit tests -
without polluting our production databases.

## Example Data Loading Function

In the above example we had a file called `app_test.go` containing a Go unit test that calls the SQL
function `app.init_data()` to initialise the data that will be used in the unit test.

The `app.init_data()` function can be created in a file called `app_test.sql` as follows:

```postgresql
-- Take care that your data loading function's name does not end in _test,
-- which would make it a unit test.
create or replace
    function app.init_data()
      returns void language 'plpgsql' as $$
    begin
        -- ...
    end;
$$;
```


If your pgpkg.toml is in the root of your Go source tree, the `app_test.go` and `app_test.sql` files will be next
to each other in the directory listing and IDE, making this mechanism exceptionally ergonomic.