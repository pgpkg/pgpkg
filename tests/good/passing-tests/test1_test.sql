create or replace function passing_tests.test_1() returns void language plpgsql as $$
    begin
        raise notice 'test 1';
    end;
$$;

create or replace function passing_tests.test_2() returns void language plpgsql as $$
    begin
        raise notice 'test 2';
    end;
$$;