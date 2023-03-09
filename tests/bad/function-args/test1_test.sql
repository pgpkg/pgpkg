--
-- Syntax error inside a table definition.
--

create or replace function function_args.test_1() returns void language plpgsql as $$
    begin
        raise notice 'test 1';
    end;
$$;

create or replace function function_args.test_2(_arg integer) returns void language plpgsql as $$
    begin
        raise notice 'test 2';
    end;
$$;