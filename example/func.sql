create or replace function hello.func() returns text language plpgsql as $$
    begin
        return 'Hello, world!';
    end;
$$;

