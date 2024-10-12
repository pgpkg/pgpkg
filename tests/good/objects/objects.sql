create or replace function "object schema".castable_to_integer(c "object schema".castable) returns integer language plpgsql as $$
begin
    return c.i;
exception
    -- handle the case where conversion fails
    when invalid_text_representation then
        raise exception 'invalid input: cannot cast % to integer', $1;
end;
$$;

create or replace
    function "object schema".t_trig()
      returns trigger language 'plpgsql' as $$
    declare
    begin
        return NEW;
    end;
$$;

comment on function "object schema".t_trig() is 'Function comments are great for postgraphile';

create cast ("object schema".castable as integer)
    with function "object schema".castable_to_integer("object schema".castable)
    as implicit;

create view "object schema".v as select * from "object schema".t;
comment on view "object schema".v is 'View comments are great for postgraphile';

create trigger trig
    after insert on "object schema".t
    for each row
    execute procedure "object schema".t_trig();

comment on column "object schema".t.t is 'Column comments are great for postgraphile';