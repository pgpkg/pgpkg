--
-- Syntax error inside a function definition. The parser seems to deal with this but
-- it dies on execution.
--

create or replace function sql_syntax.syntax(_schema text) returns void language plpgsql as $$
  begin
  end;
$$;


syntaxError;
