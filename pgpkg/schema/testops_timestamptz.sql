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

create or replace function pgpkg.timestamptz_assert_eq(_m timestamptz, _n timestamptz) returns boolean language plpgsql immutable as $$
    begin
        if _m = _n then
            return true;
        end if;

        raise exception 'assertion failed; % =? %', _m, _n;
    end;
$$;

create or replace function pgpkg.timestamptz_assert_ne(_m timestamptz, _n timestamptz) returns boolean language plpgsql immutable as $$
begin
    if _m <> _n then
        return true;
    end if;

    raise exception 'assertion failed; % <>? %', _m, _n;
end;
$$;

create or replace function pgpkg.timestamptz_assert_lt(_m timestamptz, _n timestamptz) returns boolean language plpgsql immutable as $$
begin
    if _m < _n then
        return true;
    end if;

    raise exception 'assertion failed; % <? %', _m, _n;
end;
$$;

create or replace function pgpkg.timestamptz_assert_le(_m timestamptz, _n timestamptz) returns boolean language plpgsql immutable as $$
begin
    if _m <= _n then
        return true;
    end if;

    raise exception 'assertion failed; % <=? %', _m, _n;
end;
$$;

create or replace function pgpkg.timestamptz_assert_gt(_m timestamptz, _n timestamptz) returns boolean language plpgsql immutable as $$
begin
    if _m > _n then
        return true;
    end if;

    raise exception 'assertion failed; % >? %', _m, _n;
end;
$$;

create or replace function pgpkg.timestamptz_assert_ge(_m timestamptz, _n timestamptz) returns boolean language plpgsql immutable as $$
begin
    if _m >= _n then
        return true;
    end if;

    raise exception 'assertion failed; % >=? %', _m, _n;
end;
$$;

-- drop operator if exists public.=? (timestamptz, timestamptz);
create operator public.=? (
    function = pgpkg.timestamptz_assert_eq,
    leftarg = timestamptz,
    rightarg = timestamptz
    );

-- drop operator if exists public.<>? (timestamptz, timestamptz);
create operator public.<>? (
    function = pgpkg.timestamptz_assert_ne,
    leftarg = timestamptz,
    rightarg = timestamptz
    );

-- drop operator if exists public.<? (timestamptz, timestamptz);
create operator public.<? (
    function = pgpkg.timestamptz_assert_lt,
    leftarg = timestamptz,
    rightarg = timestamptz
    );

-- drop operator if exists public.<=? (timestamptz, timestamptz);
create operator public.<=? (
    function = pgpkg.timestamptz_assert_le,
    leftarg = timestamptz,
    rightarg = timestamptz
    );

-- drop operator if exists public.>? (timestamptz, timestamptz);
create operator public.>? (
    function = pgpkg.timestamptz_assert_gt,
    leftarg = timestamptz,
    rightarg = timestamptz
    );

-- drop operator if exists public.>=? (timestamptz, timestamptz);
create operator public.>=? (
    function = pgpkg.timestamptz_assert_ge,
    leftarg = timestamptz,
    rightarg = timestamptz
    );
