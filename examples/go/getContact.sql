create or replace function example.get_contact(_contact uuid) returns text language plpgsql as $$
    begin
        return name from example.contact where contact = _contact;
    end;
$$;