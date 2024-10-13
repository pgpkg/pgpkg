--
-- Store a message after computing its hash. If the message already exists,
-- win!
--

create or replace function gl.message(_team gl.team_k, _msg text) returns uuid language plpgsql as $$
  declare
      _hash uuid = uuid_generate_v5('4B959A65-2A9A-44EF-ABF9-0CAC8597034F', _msg);

  begin
    insert into gl.message (team, hash, message) values (_team, _hash, _msg)
        on conflict do nothing;
    return _hash;
  end;
$$;

create or replace function gl.message(_team gl.team_k, _hash uuid) returns text language sql as $$
    select message from gl.message where team=_team and hash=_hash;
$$;