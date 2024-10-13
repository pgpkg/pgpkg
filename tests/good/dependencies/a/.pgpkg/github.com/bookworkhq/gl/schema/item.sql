--
-- An item represents a single line item within a transaction.
-- Items are more complex than they appear, because they also enable the recording of
-- tax, discounts and surcharges (themselves complex), and also because they are user-
-- visible and therefore need to be rounded; which must also be recorded.
--
-- Items therefore contain a list of "adjustments", which are used to store these value modifiers.
--
-- When a transaction is "applied", a list of entries is created which formalise the accounting.
-- NOTE: items are never used for financial calculations; they are used to compute the financial
-- changes necessary for a transaction to be applied to an account, and to provide a reference to
-- the goods or services for which the item applies.
--

create table gl.item (
    primary key (team, item),
    foreign key (team, tx) references gl.tx,

    team        gl.team_k         not null,
    item        gl.item_k         not null default gen_random_uuid(),
    tx          gl.tx_k           not null,
    description uuid,

    account     gl.account_k      not null,
    base        decimal           not null,          -- base value of this item before adjustments.
    adjustments gl.adjustment_t[],                   -- The bits that make up this item.
    amount      decimal           not null default 0 -- NOTE: Updated only by a trigger
);

create index item_team_tx_idx on gl.item (team, tx);