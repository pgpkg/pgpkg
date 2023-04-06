create or replace function passing_tests2.test_3() returns void language plpgsql as $$
    begin
        raise notice 'test 3';
    end;
$$;

create or replace function passing_tests2.test_4() returns void language plpgsql as $$
    begin
        raise notice 'test 4';
    end;
$$;

create or replace function passing_tests2.test_5() returns void language plpgsql as $$
begin
    raise notice 'test 5';
end;
$$;
