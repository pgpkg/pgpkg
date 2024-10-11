create or replace
    function "good.Schema Name$!".f()
      returns void language 'plpgsql' as $$
    begin
        raise notice 'it''s a good schema name; fight me';
    end;
$$;