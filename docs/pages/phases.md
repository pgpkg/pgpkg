# Installation Phases

`pgpkg` goes through a number of phases when it's installing a package. Here's a summary.

**Lock pgpkg**. Only one instance of pgpkg can run at a time.

**Create schema and roles**. The schema is created with the package's role as owner. If the owner or schema
already exist, they are assumed to belong to this package.

**Parse the API**. Bail early if there are issues with the code.

**Remove old API objects**. Functions, views and triggers that were managed by `pgpkg` are removed;
note that they can't be used during a migration.

**Grant access**. The new schema is granted access to schemas named the `Uses` clause. This means
that migrations *can* use functions, tables and other objects that they depend on.

Once this work is performed, the effective role is changed to the schema owner. As this user,
`pgpkg` will:

**Perform migration**. Run the migration scripts in order, and record for posterity.

**Install API**. Functions, views and triggers are created. Dependencies are automagically resolved.

**Run Tests**. Tests are called `test_xxx` and are run in random order.

**Commit**. If all went well, the transaction is committed.