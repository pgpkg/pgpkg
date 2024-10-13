--
-- Creates a new debtor, which is a sub-account under the `receivables` account, found in settings.
--
create or replace function gl.debtor_create(_team gl.team_k, _account_k gl.account_k, _name text) returns void language plpgsql as $$
    begin
        perform gl.account_create(_team, (gl.settings(_team)).receivables, _name, _account_k,
            2, (gl.settings(_team)).debtor_rounding, true);
    end;
$$;