--
-- This table contains de-duplicated text strings.
-- In accounting - and particularly in invoicing - there is substantial
-- duplication of string values - which bloats both the overall size of the
-- database as well as the size of individual table rows, leading to reduced
-- overall performance.
--
-- This table - and the associated functions - provide a simple hash-based deduplication
-- system.
--

create table gl.message (
    primary key (team, hash),
    
    team gl.team_k not null,
    hash uuid not null,
    message text not null
);