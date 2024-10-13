--
-- Returns the settings for a given team. This function is stable so it can be used as often as you like.
-- To access a particular setting for team #1:
--
-- # select * from gl.account where account = (gl.settings(1)).receivables;
--  team | account | currency |    parents    |        name
-- ------+---------+----------+---------------+---------------------
--     1 |   31000 |        1 | {30000,31000} | Accounts Receivable
--
-- Note the syntactic need to put the query in parens.
--
create or replace function gl.settings(_team gl.team_k) returns gl.settings stable language sql as $$
    select * from gl.settings where team = _team;
$$;