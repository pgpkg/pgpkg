create or replace function gl.test_account_test() returns void language plpgsql as $$
    begin
        raise notice 'BERK OUT!';
    end;
$$