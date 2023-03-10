create or replace function hello.world() returns text language plpgsql as $$
    declare
        _who text;

    begin
        select name into _who strict from hello.contact;
        return _who;
    end;
$$;
