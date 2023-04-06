--
-- If the search_path is set properly then we can call any function defined in the package
-- from anywhere. I don't recommend this generally, at the moment let's consider it experimental.
--

create or replace function cs1.func1() returns integer language 'sql' as $$ select 1; $$;
create or replace function cs2.func2() returns integer language 'sql' as $$ select 2; $$;

create or replace function cs1.test_cs1() returns void language plpgsql as $$
    begin
        perform func1();
        perform func2();
    end;
$$;

create or replace function cs2.test_cs2() returns void language plpgsql as $$
    begin
        perform func1();
        perform func2();
    end;
$$;