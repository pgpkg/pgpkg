--
-- Syntax error inside a function definition. The parser seems to deal with this but
-- it dies on execution.
--

create or replace function test_exception.exception() returns void language plpgsql as $$
  begin
      raise exception 'a test exception';
  end;
$$;

create or replace function test_exception.indirection() returns void language plpgsql as $$
begin
    perform test_exception.exception();
end;
$$;

select test_exception.indirection();