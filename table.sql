create extension pgcrypto;

create table users (
    id uuid default gen_random_uuid(),
    username varchar not null,
    password varchar not null,
    photo varchar not null default '',
    created_at timestamp not null default now()
);
create unique index users_username_idx on users (username);
