create or replace function hello.func() returns text language plpgsql as $$
    begin
        return 'Hello, world!';
    end;
$$;


create or replace view hello.who as select current_user;
comment on function hello.func() is 'Hello, func!';
comment on view hello.who is 'Hello, view!';
