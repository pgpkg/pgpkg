--
-- If the search_path is set properly then we can call any function defined in the package
-- from anywhere. I don't recommend this generally, at the moment let's consider it experimental.
--

create or replace function cs1.func1() returns integer language 'sql' as $$ select 1; $$;
create or replace function cs2.func2() returns integer language 'sql' as $$ select 2; $$;

create or replace function cs1.cs1_test() returns void language plpgsql as $$
    begin
        perform func1();
        perform func2();
    end;
$$;

create or replace function cs2.cs2_test() returns void language plpgsql as $$
    begin
        perform func1();
        perform func2();
    end;
$$;