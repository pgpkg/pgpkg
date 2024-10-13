--
-- Utility to create a currency value from a text string:  select core.to_currency('AUD')
--
create or replace function gl.currency(_currency text) returns gl.currency_t language sql immutable as $$
  select
    ((ascii(substring(upper(_currency), 1, 1)) - 65) * 676 +
    (ascii(substring(upper(_currency), 2, 1)) - 65) * 26 +
    (ascii(substring(upper(_currency), 3, 1)) - 65))::gl.currency_t;
$$;

--
-- Utility to print a currency as a test string:  select core.from_currency(core.to_currency('AUD')) -> 'AUD'
--
create or replace function gl.currency(_currency gl.currency_t) returns text language sql immutable as $$
  select
    chr(65 + _currency / 676) ||
    chr(65 + _currency % 676 / 26) ||
    chr(65 + _currency % 26);
$$;


--
-- Convert directly from integer to string.
--
create or replace function gl.currency(_currency integer) returns text language sql immutable as $$
    select gl.currency(_currency::gl.currency_t);
$$;