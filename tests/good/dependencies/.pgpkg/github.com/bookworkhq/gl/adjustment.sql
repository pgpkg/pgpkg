--
-- Make the adjustments on an item visible for demo purposes.
--
create or replace function gl.pretty_item_adjustments(_team_k gl.team_k, _adjustments gl.adjustment_t[]) returns text language plpgsql as $$
    declare
        _account_k gl.account_k;
        _adjustment gl.adjustment_t;
        _pretty text;

    begin
        foreach _adjustment in array _adjustments
        loop
            _pretty = concat_ws(', ', _pretty, '[' || gl.name(_team_k, _adjustment.account) || ' ' || _adjustment.amount || ']');
        end loop;

        return _pretty;
    end;
$$;