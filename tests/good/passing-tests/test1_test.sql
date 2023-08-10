create or replace function passing_tests.t1_test() returns void language plpgsql as $$
    begin
        raise notice 'test 1';
    end;
$$;

create or replace function passing_tests.t2_test() returns void language plpgsql as $$
    begin
        raise notice 'test 2';
    end;
$$;