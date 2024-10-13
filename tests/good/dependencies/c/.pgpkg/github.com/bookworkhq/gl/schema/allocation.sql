--
-- Tracks allocations between gl.entries which are registered as "unallocated"
-- (ie, in the unallocated table). This allows us to know when an invoice or other
-- transaction is "complete" from a tracking perspective.
--
-- Receivables systems generally come in two flavours: "open item" and "balance forward".
-- Of the two, open-item systems are more flexible because we can tell which transactions
-- are related, which makes debt collection and payment policies much more flexible.
-- Allocations is an implementation of an open-item system. It doesn't necessarily need to
-- be used, and is independent of the rest of the code.
--

create table gl.allocation (
    primary key (team, allocation),
    foreign key (team, credit) references gl.unallocated,
    foreign key (team, debit) references gl.unallocated,

    check (amount > 0),

    team gl.team_k not null,
    allocation gl.allocation_k not null default gen_random_uuid(),
    credit gl.entry_k not null,
    debit gl.entry_k not null,
    amount decimal not null
);

create index allocation_credit_idx on gl.allocation (team, credit);
create index allocation_debit_idx on gl.allocation (team, debit);