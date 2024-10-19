create or replace function gl.test_account_test() returns void language plpgsql as $$
    begin
        raise notice 'this is a message from gl.test_account_test()!';
    end;
$$
