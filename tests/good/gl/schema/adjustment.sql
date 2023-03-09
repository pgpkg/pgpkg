--
-- Components can be added to things like items to enumerate how a
-- transaction should be broken into accounts, before the entries are created.
-- This is important e.g. for dealing with drafts and generated charges.
--
-- Adjustments are not stored in a table, but as an array against an item. Items are
-- aggregated together to create a set of Entries which are applied to the various accounts.
--
-- Component is a type, so we don't need to record the team, because
-- it should be inferred from the object that contains it.
--
create type gl.adjustment_t as (
    account gl.account_k,
    amount decimal
);
