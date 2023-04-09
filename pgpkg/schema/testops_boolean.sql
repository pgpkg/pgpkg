create or replace function pgpkg.boolean_assert_true(_b boolean) returns boolean language plpgsql immutable as $$
begin
    if _b then
        return true;
    end if;

    raise exception 'assertion failed; ?(%)', _b;
end;
$$;

create or replace function pgpkg.boolean_assert_false(_b boolean) returns boolean language plpgsql immutable as $$
begin
    if not(_b) then
        return true;
    end if;

    raise exception 'assertion failed; ?!(%)', _b;
end;
$$;

create operator pgpkg.?? (
    function = pgpkg.boolean_assert_true,
    rightarg = boolean
    );

create operator pgpkg.?! (
    function = pgpkg.boolean_assert_false,
    rightarg = boolean
    );
