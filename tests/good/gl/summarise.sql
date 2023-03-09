--
-- summarise is a general-purpose report, intended to allow quick financial reports to be generated.
--
-- _team        the team ID
-- _currency    the reporting currency. Only entries for this currency will be displayed.
-- _depth       how many levels deep to report. Deeper levels are sum()'d
-- _periodic    break the report into periods. The periods are defined in the function gl.reporting_period,
--              which takes an entry's effective date and returns a "bucket" for summarisation.
-- _base        the root of the reporting tree. Typically settings.assets, liabilities, income or expenses.
-- _sign        +1 or -1; change the sign of values on the report. liability and income are -1, the rest are +1.
-- _not_before  only entries on or after this date are considered. Used for P&L.
-- _not_after   only entries up to or including this date are considered. Used for P&L and balance sheet history.
--
-- In `psql`, the output of summarise can be easily tabulated:
--    select account.name, reporting_period, sum(amount)
--      from gl.summarise(...)
--      join gl.account using (team, account)
--      group by account.name, reporting_period
--      order by reporting_period desc \crosstabview
--
create or replace function gl.reporting_period(_periodic boolean, _when timestamptz) returns date immutable language sql as $$
    select case _periodic when true then date_trunc('month', _when)::date end;
$$;

create or replace function gl.summarise(_team gl.team_k, _currency gl.currency_t, _depth integer, _periodic boolean, _base gl.account_k, _sign decimal, _not_before timestamptz default null, _not_after timestamptz default null) returns setof gl.report language plpgsql as $$
begin
    -- Select individual entries where the account depth is <= than requested.
    return query
      select team, array_length(account.parents, 1), gl.reporting_period(_periodic, effective_time) as reporting_period, account, account.parents, entry.currency, sum(entry.amount) * _sign
         from gl.entry join gl.account using (team, account)
         where
             (_not_before is null or entry.effective_time >= _not_before) and
             (_not_after is null or entry.effective_time <= _not_after) and
             entry.currency = _currency and
             account.parents @> array[_base] and
             array_length(account.parents, 1) <= _depth and
             account.team = _team
         group by team, entry.currency, reporting_period, account, account.parents;

    -- sum() all the entries where the account depth is > than requested.
    return query
       select team, _depth, gl.reporting_period(_periodic, effective_time) as reporting_period, account.parents[_depth] as max_acc, account.parents[:_depth] as p, entry.currency, sum(amount) * _sign -- , _depth as depth
         from gl.entry join gl.account using (team, account)
         where
             (_not_before is null or entry.effective_time >= _not_before) and
             (_not_After is null or entry.effective_time <= _not_after) and
             entry.currency = _currency and
             account.parents @> array[_base] and
             array_length(account.parents, 1) > _depth and
             account.team = _team
         group by team, entry.currency, reporting_period, max_acc, p;
end;
$$;

