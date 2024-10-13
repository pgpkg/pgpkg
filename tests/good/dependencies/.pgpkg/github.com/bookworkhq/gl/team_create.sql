--
-- Configures a new team, by setting up a complete chart of accounts for it.
-- Note: currency is going to disappear.
--
create or replace function gl.team_create(_team gl.team_k, _name text) returns gl.team_k language plpgsql as $$
    declare
        -- Income accounts
        ACIncome gl.account_k = 10000;
        ACSales gl.account_k = 11000;
        ACSurcharges gl.account_k = 12000;
        ACRounding gl.account_k = 13000;
        ACSalesRounding gl.account_k = 13001;
        ACTaxRounding gl.account_k = 13002;
        ACDiscountRounding gl.account_k = 13003;
        ACItemRounding gl.account_k = 13004;
        ACDebtorRounding gl.account_k = 13005;
        ACCredits gl.account_k = 14000;
        ACDebits gl.account_k = 15000;

        -- Expense accounts
        ACExpenses gl.account_k = 20000;
        ACDiscounts gl.account_k = 21000;
        ACCOGS gl.account_k = 22000;

        -- Assets
        ACAssets gl.account_k = 30000;
        ACReceivables gl.account_k = 31000;
        ACStripe gl.account_k = 32000;
        ACBank gl.account_k = 33000;

        -- Liabilities
        ACLiabilities gl.account_k = 40000;
        ACTax gl.account_k = 41000;

    begin
        insert into gl.team (team, name) values (_team, _name);

        insert into gl.account (team, account, parents, name)
            values (_team, ACIncome, array[ACIncome], 'Income');

        insert into gl.account (team, account, parents, name)
            values (_team, ACSales, array[ACIncome, ACSales], 'Sales');

        insert into gl.account (team, account, parents, name)
            values (_team, ACSurcharges, array[ACIncome, ACSurcharges ], 'Surcharges');

        insert into gl.account (team, account, parents, name)
            values (_team, ACRounding, array[ACAssets, ACRounding], 'Rounding');

        insert into gl.account (team, account, parents, name)
            values (_team, ACSalesRounding, array[ACAssets, ACRounding, ACSalesRounding], 'Sales Rounding');

        insert into gl.account (team, account, parents, name)
            values (_team, ACItemRounding, array[ACAssets, ACRounding, ACItemRounding], 'Item Rounding');

        insert into gl.account (team, account, parents, name)
            values (_team, ACDebtorRounding, array[ACAssets, ACRounding, ACDebtorRounding], 'Debtor Rounding');

        insert into gl.account (team, account, parents, name)
            values (_team, ACTaxRounding, array[ACAssets, ACRounding, ACTaxRounding], 'Tax Rounding');

        insert into gl.account (team, account, parents, name)
            values (_team, ACDiscountRounding, array[ACAssets, ACRounding, ACDiscountRounding], 'Discount Rounding');

        insert into gl.account (team, account, parents, name)
            values (_team, ACCredits, array[ACIncome, ACCredits], 'Credit Notes');

        insert into gl.account (team, account, parents, name)
            values (_team, ACDebits, array[ACIncome, ACDebits], 'Debit Notes');

        insert into gl.account (team, account, parents, name)
            values (_team, ACExpenses, array[ACExpenses], 'Expenses');

        insert into gl.account (team, account, parents, name, rounding_dps, rounding_account)
            values (_team, ACDiscounts, array[ACExpenses, ACDiscounts], 'Discounts given', 2, ACDiscountRounding);

        insert into gl.account (team, account, parents, name)
            values (_team, ACCOGS, array[ACExpenses, ACCOGS], 'Cost of sales');


--         insert into gl.account (team, account, parents, name)
--             values (_team, 11100, array[ACIncome, ACSales, 11100], 'Widgets');
--
--         insert into gl.account (team, account, parents, name)
--             values
--                 (_team, 11110, array[ACIncome, ACSales, 11100, 11110], 'Bidgets'),
--                 (_team, 11111, array[ACIncome, ACSales, 11100, 11111], 'Didgets'),
--                 (_team, 11112, array[ACIncome, ACSales, 11100, 11112], 'Fidgets'),
--                 (_team, 11113, array[ACIncome, ACSales, 11100, 11113], 'Gidgets'),
--                 (_team, 11114, array[ACIncome, ACSales, 11100, 11114], 'Hidgets'),
--                 (_team, 11115, array[ACIncome, ACSales, 11100, 11115], 'Kidgets'),
--                 (_team, 11116, array[ACIncome, ACSales, 11100, 11116], 'Lidgets'),
--                 (_team, 11117, array[ACIncome, ACSales, 11100, 11117], 'Midgets'),
--                 (_team, 11118, array[ACIncome, ACSales, 11100, 11118], 'Nidgets'),
--                 (_team, 11119, array[ACIncome, ACSales, 11100, 11119], 'Pidgets');
--
--         insert into gl.account (team, account, parents, name)
--             values (_team, 11200, array[ACSales, 11200], 'Doovers');
--
--         insert into gl.account (team, account, parents, name)
--             values
--                 (_team, 11210, array[ACIncome, ACSales, 11200, 11210], 'Boovers'),
--                 (_team, 11211, array[ACIncome, ACSales, 11200, 11211], 'Coovers'),
--                 (_team, 11212, array[ACIncome, ACSales, 11200, 11212], 'Foovers'),
--                 (_team, 11213, array[ACIncome, ACSales, 11200, 11213], 'Goovers'),
--                 (_team, 11214, array[ACIncome, ACSales, 11200, 11214], 'Hoovers'),
--                 (_team, 11215, array[ACIncome, ACSales, 11200, 11215], 'Joovers'),
--                 (_team, 11216, array[ACIncome, ACSales, 11200, 11216], 'Koovers'),
--                 (_team, 11217, array[ACIncome, ACSales, 11200, 11217], 'Loovers'),
--                 (_team, 11218, array[ACIncome, ACSales, 11200, 11218], 'Moovers'),
--                 (_team, 11219, array[ACIncome, ACSales, 11200, 11219], 'Noovers');


        insert into gl.account (team, account, parents, name)
            values (_team, ACAssets, array[ACAssets], 'Assets');

        insert into gl.account (team, account, parents, name)
            values (_team, ACReceivables, array[ACAssets, ACReceivables], 'Accounts Receivable');

        insert into gl.account (team, account, parents, name)
            values (_team, ACStripe, array[ACAssets, ACStripe], 'Stripe');

        insert into gl.account (team, account, parents, name)
            values (_team, ACBank, array[ACAssets, ACBank], 'Bank Account');

        insert into gl.account (team, account, parents, name)
            values (_team, ACLiabilities, array[ACLiabilities], 'Liabilities');

        insert into gl.account (team, account, parents, name)
            values (_team, ACTax, array[ACLiabilities, ACTax], 'Tax');

        -- WARNING: this default tax code is used by lotus and should not be deleted.
        insert into gl.account (team, account, parents, name, rounding_dps, rounding_account)
            values (_team, 41100, array[ACLiabilities, ACTax, 41100], 'Sales Tax Payable', 2, ACTaxRounding);

        insert into gl.account (team, account, parents, name, rounding_dps, rounding_account)
            values (_team, 41200, array[ACLiabilities, ACTax, 41200], 'PST Tax Payable', 2, ACTaxRounding);

        insert into gl.account (team, account, parents, name, rounding_dps, rounding_account)
            values (_team, 41300, array[ACLiabilities, ACTax, 41300], 'HST Tax Payable', 2, ACTaxRounding);


        insert into gl.settings
            (team, reporting_timezone, reporting_cycle_start, reporting_cutoff,
             assets, liabilities, tax, income, sales, expenses,
             receivables, credits, debits, item_rounding, debtor_rounding)
            values (1, 'Australia/Melbourne', '2000-7-1', '2000-1-1',
                    ACAssets, ACLiabilities, ACTax, ACIncome, ACSales, ACExpenses,
                    ACReceivables, ACCredits, ACDebits, ACItemRounding, ACDebtorRounding);

        return _team;
    end;
$$;