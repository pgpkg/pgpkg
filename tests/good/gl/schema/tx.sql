--
-- A transaction is just a list of entries. It allows us to find related entries
-- if we need to audit or otherwise make changes. Any entry must have an associated
-- transaction.
--
-- It's likely we will want to add some audit information to the transaction in
-- the future. However, in order to make reporting easy, any information in a
-- transaction needs to be copied to the individual entries anyway.
--
-- The "amount" of a transaction is the transaction amount relative to the purpose
-- of the transaction; for example, an invoice tx for "100" means that the invoice value
-- is $100, and the items of the invoice will add up to 100. Note however that the GL entries
-- for an invoice - or any other transaction - will always add up to zero.
--
-- Note that the "effective time" of a transaction may be modified if a transaction is finalised
-- past the cutoff date for reporting.
--
create table gl.tx (
    primary key (team, tx),

    team    gl.team_k not null,
    tx      gl.tx_k   not null default gen_random_uuid(),
    tx_type gl.tx_type_e not null,

    currency       gl.currency_t not null,
    created_time   timestamptz        not null default current_timestamp, -- audit flag
    effective_time timestamptz        not null default current_timestamp, -- proposed transaction date

    -- these values are not set until the tx is finalised.
    amount         decimal,
    finalised_time timestamptz
);