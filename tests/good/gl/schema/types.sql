--
-- Giving names to data types help us understand what they do and
-- what they relate to.
--
-- I'm using the following convention:
--
--     name       is a table
--     name_t     is a compound data type (adjustment_t) or a primitive type not otherwise declared (currency_t).
--     name_k     is the type of the key column for the table of the same name (entry_k).
--     name_e     is an enumeration type
--
-- similar names are also used in the stored functions. For example, in pl/PGSQL:
--
--    gl.entry     refers to a row of the `entry` table.
--    gl.entry_k   refers to the primary key of the `entry` table.
--
-- NOTE: most of these types should probably be uuid, and this requires review.
--
create domain gl.account_k as integer;
create domain gl.team_k as integer;
create domain gl.item_k as uuid;
create domain gl.tx_k  as uuid;
create domain gl.entry_k as uuid;
create domain gl.allocation_k as uuid;
create domain gl.currency_t as smallint;

create type gl.tx_type_e as enum ('invoice', 'receipt', 'debit', 'credit');
create type gl.account_type_e as enum ('asset', 'liability', 'expense', 'revenue', 'equity');

-- Used for financial reporting (surprise!)
create type gl.report as (
    team gl.team_k,
    depth integer,
    reporting_period date,
    account gl.account_k,
    parents gl.account_k[],
    currency gl.currency_t,
    amount decimal
);
