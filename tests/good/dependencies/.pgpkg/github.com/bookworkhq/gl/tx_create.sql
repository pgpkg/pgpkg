--
-- Create a tx record and return its ID.
--
create or replace function gl.tx_create(_team gl.team_k, _tx_type gl.tx_type_e, _currency gl.currency_t, _effective_time timestamptz) returns gl.tx_k language sql as $$
    insert into gl.tx (team, tx_type, currency, effective_time) values (_team, _tx_type, _currency, _effective_time) returning tx;
$$;