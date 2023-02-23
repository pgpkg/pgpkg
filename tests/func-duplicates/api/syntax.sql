--
-- Syntax error inside a function definition. The parser seems to deal with this but
-- it dies on execution.
--

create or replace function func_duplicates.f1() returns void language plpgsql as $$
  begin
      raise notice 'the first function';
  end;
$$;

create or replace function func_duplicates.f1() returns void language plpgsql as $$
    begin
        raise notice 'the second function';
    end;
$$;