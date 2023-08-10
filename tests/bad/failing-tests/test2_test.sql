--
-- Syntax error inside a table definition.
--

create or replace function failing_tests.t3_test() returns void language plpgsql as $$
    begin
        raise notice 'test 3';
    end;
$$;

create or replace function failing_tests.t4_test() returns void language plpgsql as $$
    begin
        raise notice 'test 4';
    end;
$$;

create or replace function failing_tests.t5_test() returns void language plpgsql as $$
begin
    raise exception 'this test is expected to fail';
end;
$$;