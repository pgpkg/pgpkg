--
-- This trigger maintains the 'amount' column in an item, by updating it when adjustments
-- are added. It also ensures that all items are rounded to 2dps, by adding a rounding
-- adjustment if needed.
--
create or replace function gl.item_amount_trigger() returns trigger language plpgsql as $$
    declare
        _rounding_k gl.account_k = (gl.settings(NEW.team)).item_rounding;
        _rounding decimal;
        _amount decimal;

    begin
        -- Add up all the adjustments to date.
        _amount = coalesce((select sum(amount) from unnest(NEW.adjustments)), 0) + NEW.base;

        -- compute a rounding adjustment to bring the line item value to 2dp.
        _rounding = round(_amount, 2) - _amount;
        if _rounding <> 0 then
            NEW.adjustments = NEW.adjustments || (_rounding_k, _rounding)::gl.adjustment_t;
            _amount = _amount + _rounding;
        end if;

        NEW.amount = round(_amount, 2);
        return NEW;
    end;
$$;

create or replace trigger amount_trigger
    before insert or update of base, adjustments, amount
    on gl.item
    for each row
    execute procedure gl.item_amount_trigger();
