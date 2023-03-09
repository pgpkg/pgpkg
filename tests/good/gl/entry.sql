--
-- This trigger ensures that the entry currency always matches the tx currency.
-- It is critical to ensure that entry currencies do not vary from the tx.
--
create or replace function gl.entry_trigger() returns trigger language plpgsql as $$
    declare
        _currency gl.currency_t;

    begin
        _currency = currency from gl.tx where team=NEW.team and tx=NEW.tx;
        if _currency <> NEW.currency then
            raise exception 'entry % currency % does not match tx % currency %', entry.entry, entry.currency, entry.tx, _currency;
        end if;

        return NEW;
    end;
$$;

create or replace trigger currency_trigger
    before insert or update
    on gl.entry
    for each row
    execute procedure gl.entry_trigger();
