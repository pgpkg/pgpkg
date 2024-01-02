create table example.contact (
    primary key (contact),
    contact uuid not null default gen_random_uuid(),
    name text
);

insert into example.contact (contact, name) values ('90069E0B-2998-45E0-B8FC-91610761B429', 'Mark Lillywhite');
