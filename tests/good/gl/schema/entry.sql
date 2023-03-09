--
-- Entries in an account.
-- A single transaction consists of multiple entries. The entries determine how the
-- account is managed. For example, when an invoice is finalised, the value is split
-- into at least three accounts:
--
--   * Sales
--   * Tax withheld
--   * Accounts Receivable
--
-- The sum of all entries in a transaction, and for the whole database, should **always**
-- and **without exception** equal zero. A non-zero sum at any level means there's a bug.
-- That's really the point of double-entry systems like this.
--

create table gl.entry (
    primary key (team, entry),
    foreign key (team, account) references gl.account,
    foreign key (team, tx) references gl.tx,

    team           gl.team_k     not null,
    entry          gl.entry_k    not null default gen_random_uuid(),
    tx             gl.tx_k       not null,
    account        gl.account_k  not null,
    effective_time timestamptz   not null default current_timestamp,
    currency       gl.currency_t not null,
    amount         decimal       not null
);

create index entry_team_account_idx on gl.entry (team, account);

create index entry_team_tx_idx on gl.entry (team, tx);