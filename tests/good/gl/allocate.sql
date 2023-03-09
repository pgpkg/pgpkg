--
-- Record funds allocation between two entries. Both entries must already have an
-- unallocated table entry.
--
-- Naturally, allocation is only permitted between two entries of the same currency.
--

create or replace function gl.allocate(_team gl.team_k, _credit_k gl.entry_k, _debit_k gl.entry_k, _amount decimal) returns void language plpgsql as $$
declare
    _credit gl.entry;
    _debit gl.entry;
    _cunalloc gl.unallocated;
    _dunalloc gl.unallocated;


begin
    select * into strict _credit from gl.entry where team=_team and entry=_credit_k;
    select * into strict _cunalloc from gl.unallocated where team=_team and entry=_credit_k;
    select * into strict _debit from gl.entry where team=_team and entry=_debit_k;
    select * into strict _dunalloc from gl.unallocated where team=_team and entry=_debit_k;

    if _amount <= 0 then
        raise exception 'allocation amount % must be greater than zero', _amount;
    end if;

    if _credit.currency <> _debit.currency then
        raise exception 'Unable to allocate credit % to debit %: different currencies', _credit.entry, _debit.entry;
    end if;

    if _dunalloc.amount < 0 then
        raise exception 'Entry % is not a debit', _debit;
    end if;

    if _cunalloc.amount > 0 then
        raise exception 'Entry % is not a credit', _credit;
    end if;

    if _amount > _dunalloc.amount then
        raise exception 'Unable to allocate % to debit %; only % unallocated', _amount, _debit.entry, _dunalloc.amount;
    end if;

    if -_amount < _cunalloc.amount then
        raise exception 'Unable to allocate % to debit %; only % unallocated', _amount, _credit.entry, _cunalloc.amount;
    end if;

    insert into gl.allocation (team, credit, debit, amount)
        values (_team, _credit_k, _debit_k, _amount);

    update gl.unallocated set amount = amount - _amount where team=_team and entry=_debit_k;
    update gl.unallocated set amount = amount + _amount where team=_team and entry=_credit_k;
end;
$$;