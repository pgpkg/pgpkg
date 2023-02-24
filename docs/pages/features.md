# pgpkg features

pgpkg eliminates the need to write migration scripts for functions, views  and triggers.
It does this by treating these objects as if they were mutable, in the same way that a
function or method in a Go, Java or Python program is mutable - you can just change it.
This is a different approach to most migration tools which treat functions, views and
triggers in the same way they treat tables and user-defined types.

## Parser Based

pgpkg uses Postgres' own parser to parse the SQL scripts defined in a package. This means
that pgpkg is able to understand function, view and trigger declarations, and it uses this
understanding to manage these objects.

In the future, the hope is to continue to extend the use of the parser to enable greater safety,
as well as the ability to implement features such as schema rewriting.

## 

parser based
raise notice
contextual errors
transactional safety
non destructive tests

rollback scripts NOT