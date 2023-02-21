--
-- An API consists of the complete list of functions, views and triggers
-- created by pgpkg. pgpkg remembers the order in which it successfully created
-- API objects, as an optimisation to improve build and delete times.
--
-- Note that the order is not hard; pgpkg will try to create and delete all objects
-- in an API bundle, even if the stored order results in dependency errors.
-- This could certainly happen when new objects are added.
--
create table pgpkg.api (
    primary key (pkg, obj_type, obj_name),

    -- name of the package
    pkg text,

    -- the order of successful creation.
    seq integer not null,

    -- One of "function", "view" or "trigger"
    obj_type text,

    -- full name of the object, suitable for DROP FUNCTION, DROP VIEW etc..
    obj_name text
);