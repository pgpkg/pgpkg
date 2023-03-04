create or replace function hello.test_world() returns void language plpgsql as $$
  begin
    if hello.world() <> 'Postgresql Community' then
      raise exception 'the world is not right';
    end if;
  end;
$$;
