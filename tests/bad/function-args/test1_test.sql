--
-- This test definition is OK.
--
create or replace function function_args.t1_test() returns void language plpgsql as $$
    begin
        raise notice 'test 1';
    end;
$$;

--
-- A test should not have arguments, and won't be run. So we don't allow it.
--
create or replace function function_args.t2_test(_arg integer) returns void language plpgsql as $$
    begin
        raise notice 'test 2';
    end;
$$;