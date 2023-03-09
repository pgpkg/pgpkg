--
-- This table tracks the unallocated value of an accounting entry, such as
-- a credit or debit. Entries in this table are created automatically by the
-- high-level functions such as invoice, receipt, credit, debit.
--
-- Note: not all entries have allocations. We only create an allocation table
-- entry for entries that need following up.
--
create table gl.unallocated (
    primary key (team, entry),

    team gl.team_k not null,
    entry gl.entry_k not null,
    amount decimal not null
);
