--
-- An account is line on a report, such as a P&L or ledger.
-- Accounts represent detailed areas of the business, down to
-- the level of individual debtors and creditors.
--
-- Accounts can be grouped hierarchically. To make queries super
-- fast, we record the hierarchy on each account. This allows us to
-- quickly sum subcategories of a report without recursive SQL.
--
create table gl.account (
    primary key (team, account),

    check (parents[array_upper(parents, 1)] = account),
    check ((rounding_dps is null) = (rounding_account is null)),

    team             gl.team_k      not null,
    account          gl.account_k   not null,
    parents          gl.account_k[] not null,               -- must include itself at least
    name             text           not null,               -- should be a hash
    open_item        boolean        not null default false, -- create an allocation for entries?
    rounding_dps     smallint,                              -- should entries be rounded?
    rounding_account gl.account_k                           -- where the rounding goes
);

