--
-- Registry of package names and schemas.
--
-- FIXME: this

create table pgpkg.pkg (
    primary key (pkg),

    pkg          text   not null,
    schema_names text[] not null,
    uses         text[]
);