--
-- Simple function to add a credit to an account (party).
-- Since this is a double-entry system, a credit to a debtor means we have to perform a debit somewhere.
-- The account that is debited is defined by settings.credit.
--
create or replace function gl.apply_credit(_team_k gl.team_k, _account_k gl.account_k, _currency gl.currency_t, _effective_time timestamptz, _amount decimal)
  returns gl.tx_k language plpgsql as $$
    declare
        _tx uuid;
        _entry gl.entry_k;

    begin
        insert into gl.tx (team, tx_type) values (_team_k, 'credit') returning tx into _tx;

        -- debit the credits account
        insert into gl.entry(team, tx, account, effective_time, currency, amount)
            values (_team_k, _tx, (gl.settings(_team_k)).credits, _effective_time, _currency, _amount);

        -- credit the receivables account.
        insert into gl.entry(team, tx, account, effective_time, currency, amount)
            values (_team_k, _tx, _account_k, _effective_time, _currency, -_amount)
            returning entry into _entry;

        -- credits need to be allocated.
        insert into gl.unallocated (team, entry, amount) values (_team_k, _entry, -_amount);

        return _tx;
    end;
$$;