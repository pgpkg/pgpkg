# Dependencies todo

- [ ] we need an ordered package group or collection/tree of packages and deps
    - [ ] this is also where versioning would become reified
        - https://go.dev/doc/modules/version-numbers
        - https://research.swtch.com/vgo-mvs

major version changes require their own schema (hello_v2) - so underscores are not allowed in pgpkg schema names

should the schema name refer to the *developer* rather than the package ?! if so then we have more problems. 

we work similarly to go in terms of resolving dependencies, but one important difference is that it is never possible
to downgrade a dependency.

