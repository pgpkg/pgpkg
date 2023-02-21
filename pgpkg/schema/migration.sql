--
-- Keep track of the migrations which have already been run for a package.
--

create table pgpkg.migration (
    -- name of the package
    pkg text,

    -- path of the migration script
    path text
);