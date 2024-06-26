--
-- These operators need to be installed using migration scripts because we don't currently
-- support operators in MOBs, and the operators need to be in the public schema.
--

--
-- This is a set of assertion operators you can use when writing tests.
--
-- For example, to assert the expected value of 120 in a test, you'd write
--
--   perform sum(amount) =? 120;
--
-- These assertion operators either return true, or throw an exception. This means that many
-- assertions can be made in a single statement. See op_test.sql for examples.
--

create or replace function pgpkg.jsonb_assert_eq(_m jsonb, _n jsonb) returns boolean language plpgsql immutable as $$
    begin
        if _m = _n then
            return true;
        end if;

        raise exception 'assertion failed; % =? %', _m, _n;
    end;
$$;

create or replace function pgpkg.jsonb_assert_ne(_m jsonb, _n jsonb) returns boolean language plpgsql immutable as $$
begin
    if _m <> _n then
        return true;
    end if;

    raise exception 'assertion failed; % <>? %', _m, _n;
end;
$$;

create operator pgpkg.=? (
    function = pgpkg.jsonb_assert_eq,
    leftarg = jsonb,
    rightarg = jsonb
    );

create operator pgpkg.<>? (
    function = pgpkg.jsonb_assert_ne,
    leftarg = jsonb,
    rightarg = jsonb
    );
