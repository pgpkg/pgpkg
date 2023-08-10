create or replace function passing_tests2.t3_test() returns void language plpgsql as $$
    begin
        raise notice 'test 3';
    end;
$$;

create or replace function passing_tests2.t4_test() returns void language plpgsql as $$
    begin
        raise notice 'test 4';
    end;
$$;

create or replace function passing_tests2.t5_test() returns void language plpgsql as $$
begin
    raise notice 'test 5';
end;
$$;
