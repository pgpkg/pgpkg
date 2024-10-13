--
-- This is a statement view.
--

create or replace function gl.age(_time timestamptz) returns integer language sql as $$
  select round(extract(day from now()::timestamp - _time) / 30, 0) * 30;
$$;

create or replace view gl.statement as
    select team, entry.entry, account, entry.effective_time, gl.age(entry.effective_time) as age, tx.tx_type, entry.currency, entry.amount, unallocated.amount as unallocated
        from gl.entry
            join gl.account using (team, account)
            join gl.tx using (team, tx)
            join gl.unallocated using (team, entry)
       order by effective_time desc;

create or replace view gl.aged_balance as
    select team, account, least(age, 120) as age, max(age) as max_age, currency, count(*) as count, sum(unallocated) as amount
      from gl.statement
      group by team, account, least(age, 120), currency
      order by age;