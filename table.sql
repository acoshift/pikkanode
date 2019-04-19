create extension pgcrypto;

create table users (
    id         uuid               default gen_random_uuid(),
    username   varchar   not null,
    password   varchar   not null,
    photo      varchar   not null default '',
    created_at timestamp not null default now(),
    primary key (id)
);
create unique index users_username_idx on users (username);

create table pictures (
    id         bigserial,
    user_id    uuid      not null,
    name       varchar   not null,
    detail     varchar   not null default '',
    photo      varchar   not null,
    tags       varchar[] not null default '{}',
    created_at timestamp not null default now(),
    primary key (id),
    foreign key (user_id) references users
);
create index on pictures (created_at desc);
create index on pictures (user_id, created_at desc);

create table favorites (
    user_id    uuid,
    picture_id bigint,
    created_at timestamp not null default now(),
    primary key (user_id, picture_id),
    foreign key (user_id) references users (id),
    foreign key (picture_id) references pictures (id)
);
