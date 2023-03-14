create or replace function pgpkg.test_ops() returns void language plpgsql as $$
    declare
        _now timestamptz = current_timestamp;
        _before timestamptz = _now - '1 day'::interval;
        _after timestamptz = _now + '1 day'::interval;

    begin
        perform 1 =? 1 and 2 <? 3 and 2 <=? 3 and 3 <=? 3 and 4 >=? 4 and 5 >=? 4 and 6 >? 5 and 6 <>? 7;
        perform 1.0 =? 1.0 and 2.9 <? 3.0 and 2.9 <=? 3.0 and 3.0 <=? 3.0 and 4.0 >=? 4.0 and 5.0 >=? 4.9 and 6.0 >? 5.9 and 6.9 <>? 7.0;
        perform _now =? _now and _before <? _now and _before <=? _now and _now <=? _now and _now >=? _before and _now >=? _now and _after >? _before and _before <>? _after;
        perform 'text' =? 'text' and 'text' <>? 'texta';
    end;
$$;