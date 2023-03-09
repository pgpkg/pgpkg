--
-- Add a new account, based on an existing one.
--
-- FIXME: key generation in this function is pretty broken. You won't add many accounts
--  before you hit duplicate keys. This will be fixed later.
--

create or replace function gl.account_create(_team_k gl.team_k, _parent_k gl.account_k, _name text, _account_k gl.account_k default null, _rounding_dps integer default null, _rounding_account gl.account_k default null, _open_item boolean default false)
    returns gl.account_k language plpgsql as $$
    declare
        _parent gl.account;

    begin
        select * into strict _parent from gl.account where team=_team_k and account=_parent_k;

        -- FIXME: this is a very slow and stupid way to create a new account identifier.
        if _account_k is null then
            _account_k = max(account)+1 from gl.account where team=_team_k and _parent.parents <@ account.parents;
        end if;

        insert into gl.account (team, account, parents, name, rounding_dps, rounding_account, open_item)
          values (_team_k, _account_k, _parent.parents || _account_k, _name, _rounding_dps, _rounding_account, _open_item);

        return _account_k;
    end;
$$;

--
-- Print a nice representation of an account, showing the parents.
--
create or replace function gl.path(_team gl.team_k, _account_k gl.account_k) returns text language plpgsql as $$
    declare
        _account gl.account;
        _parent gl.account_k;
        _name text;
        _path text;

    begin
        select * into strict _account from gl.account where team=_team and account=_account_k;
        foreach _parent in array _account.parents
        loop
            select name into strict _name from gl.account where team=_team and account=_parent;
            _path = concat_ws(' / ', _path, _name);
        end loop;

        return _path;
    end;
$$;

create or replace function gl.name(_team gl.team_k, _account_k gl.account_k) returns text language sql as $$
  select name from gl.account where team=_team and account=_account_k;
$$;