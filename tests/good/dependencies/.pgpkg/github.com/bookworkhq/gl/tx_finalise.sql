--
-- Finalise a transaction.
-- This takes the items and their adjustments and generates a bunch of Entries.
--

create or replace function gl.tx_finalise(_team_k gl.team_k, _tx_k gl.tx_k, _account_k gl.account_k) returns void language plpgsql as $$
    declare
        _tx gl.tx;
        _item gl.item;
        _adjustment gl.adjustment_t;
        _item_adjustments gl.adjustment_t[];
        _entries gl.adjustment_t[];
        _item_amount decimal = 0;
        _tx_amount decimal = 0;
        _rounding_dps smallint;
        _rounding_account gl.account_k;
        _rounding decimal;
        _entry_k gl.entry_k;
        _open_item boolean;

    begin
        select * into strict _tx from gl.tx where team=_team_k and tx=_tx_k for update;

        --
        -- Create a list of all of the item adjustments.
        -- This includes the original item values along with
        -- the adjustments. Grouping them together like this
        -- makes it easier to group and update the ledger entries.
        --
        for _item in select * from gl.item where team=_team_k and tx=_tx_k for update
        loop
            if array_length(_item.adjustments, 1) > 0 then
                -- Get total value of adjustments to the base amount, if any.
                _item_amount = _item.base + (select sum(amount) from unnest(_item.adjustments));
            else
                _item_amount = _item.base;
            end if;
            _tx_amount = _tx_amount + _item_amount;

            -- CAREFUL: _item_adjustments is the adjustments for all items. _item.adjustments is the individual item.
            _item_adjustments = _item_adjustments || _item.adjustments || (_item.account, _item.base)::gl.adjustment_t ;
        end loop;

        -- create an "adjustment" for the actual amount of the transaction.
        _item_adjustments = _item_adjustments || (_account_k, -_tx_amount)::gl.adjustment_t ;

        --
        -- Perform any necessary rounding on the item adjustments before applying them to
        -- accounts as entries. What we're doing here is creating the entire set of entries
        -- that we want to create for the tx, which we will then aggregate when creating the
        -- actual entries.
        --
        for _adjustment in select account, sum(amount) from unnest(_item_adjustments) group by account order by account
        loop
            select rounding_dps, rounding_account into strict _rounding_dps, _rounding_account from gl.account where team=_team_k and account=_adjustment.account;
            if _rounding_dps is not null then
                _rounding = _adjustment.amount - round(_adjustment.amount, _rounding_dps);
                if _rounding <> 0 then
                    _entries = _entries || (_rounding_account, _rounding)::gl.adjustment_t;
                    _adjustment.amount = round(_adjustment.amount - _rounding, _rounding_dps);
                else
                    _adjustment.amount = round(_adjustment.amount, _rounding_dps);
                end if;
            end if;
            _entries = _entries || _adjustment;
        end loop;

        -- Summarise the entries by account, and create gl.entries against the tx.
        for _adjustment in select account, sum(amount) from unnest(_entries) group by account
        loop
            if _adjustment.amount is not null then
                insert into gl.entry (team, tx, effective_time, account, currency, amount)
                    values (_team_k, _tx_k, _tx.effective_time, _adjustment.account, _tx.currency, -_adjustment.amount)
                    returning entry into _entry_k;

                select open_item into _open_item from gl.account where team=_team_k and account=_adjustment.account;
                if _open_item then
                    insert into gl.unallocated (team, entry, amount) values (_team_k, _entry_k, -_adjustment.amount);
                end if;

            end if;
        end loop;

        update gl.tx set finalised_time=current_timestamp, amount=_tx_amount where team=_team_k and tx=_tx_k;
    end;
$$;

--
-- Useful function for debugging things as they happen.
--
create or replace function gl.print_adjustments(_team_k gl.team_k, _msg text, _adjustments gl.adjustment_t[]) returns void language plpgsql as $$
    declare
        _adjustment gl.adjustment_t;
        _name text;
        _total decimal = 0;

    begin
        raise notice '--';
        for _adjustment in select * from unnest(_adjustments) order by account
            loop
            _name = name from gl.account where team=_team_k and account=_adjustment.account;
            raise notice '%: adjustment [%] % amount %', _msg, _adjustment.account, rpad(_name, 20), _adjustment.amount;
            _total = _total + _adjustment.amount;
        end loop;
        raise notice 'total adjustments %', _total;
    end;
$$;

--
-- Useful function for debugging things as they happen.
--
create or replace function gl.print_entries(_team_k gl.team_k, _msg text, _tx_k gl.tx_k) returns void language plpgsql as $$
    declare
        _entry gl.entry;
        _name text;
        _total decimal = 0;
    begin
        raise notice '--';
        for _entry in select * from gl.entry where team=_team_k and tx=_tx_k order by account
        loop
            _name = name from gl.account where team=_team_k and account=_entry.account;
            raise notice '%: entry [%] % amount %', _msg, _entry.account, rpad(_name, 20), _entry.amount;
            _total = _total + _entry.amount;
        end loop;
        raise notice 'total adjustments %', _total;
    end;
$$;