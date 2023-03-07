--
-- Managed objects are the complete list of functions, views and triggers
-- created by pgpkg.
--
create table pgpkg.managed_object (
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