--
-- Team configuration. This doesn't do much at the moment.
--

create table gl.team (
    primary key (team),

    team gl.team_k,
    name text
);