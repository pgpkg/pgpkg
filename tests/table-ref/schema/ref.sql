--
-- Missing reference in a table
--
create table table_ref.ok_table (
    ok_table integer
);

create table table_ref.bad_table (
    field integer not null,
    ref integer not null references some_reference (ref)
);