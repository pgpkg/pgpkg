--
-- Contains a list of settings for each team. Most importantly, provides the IDs
-- for accounts that are used to manage receivables.
--
create table gl.settings (
    team                  gl.team_k    not null primary key,
    foreign key (team, assets) references gl.account,
    foreign key (team, liabilities) references gl.account,
    foreign key (team, income) references gl.account,
    foreign key (team, expenses) references gl.account,
    foreign key (team, receivables) references gl.account,
    foreign key (team, credits) references gl.account,
    foreign key (team, debits) references gl.account,

    -- reporting_timezone is the timezone used when categorising reports into e.g. months
    reporting_timezone    text         not null default 'Australia/Brisbane',

    -- the basis for the first day of the financial year for this team.
    reporting_cycle_start date         not null default '2000-7-1',

    -- financial entries are only made on or after this date.
    -- needed for reporting stability.
    reporting_cutoff      timestamptz  not null,

    assets                gl.account_k not null,
    liabilities           gl.account_k not null,
    tax                   gl.account_k not null,
    income                gl.account_k not null,
    sales                 gl.account_k not null,
    expenses              gl.account_k not null,
    receivables           gl.account_k not null,
    credits               gl.account_k not null, -- where credit notes come from (income account).
    debits                gl.account_k not null, -- where debit notes come from (income account).
    item_rounding         gl.account_k not null, -- where line items get rounded into.
    debtor_rounding       gl.account_k not null  -- where we send any needed debtor rounding (shouldn't ever need this)
);