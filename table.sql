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

create table works (
    id         bigserial,
    user_id    uuid      not null,
    name       varchar   not null,
    detail     varchar   not null default '',
    photo      varchar   not null,
    tags       varchar[] not null default '{}',
    created_at timestamp not null default now(),
    primary key (id),
    foreign key (user_id) references users on delete cascade
);
create index on works (created_at desc);
create index on works (user_id, created_at desc);

create table favorites (
    user_id    uuid,
    work_id bigint,
    created_at timestamp not null default now(),
    primary key (user_id, work_id),
    foreign key (user_id) references users (id) on delete cascade,
    foreign key (work_id) references works (id) on delete cascade
);

create table comments (
    id         uuid               default gen_random_uuid(),
    work_id bigint    not null,
    user_id    uuid      not null,
    content    varchar   not null,
    created_at timestamp not null default now(),
    primary key (id),
    foreign key (work_id) references works (id) on delete cascade,
    foreign key (user_id) references users (id) on delete cascade
);
create index on comments (work_id, created_at desc);

create table follows (
    user_id      uuid,
    following_id uuid,
    created_at   timestamp not null default now(),
    primary key (user_id, following_id),
    foreign key (user_id) references users (id) on delete cascade,
    foreign key (following_id) references users (id) on delete cascade
);
create index on follows (user_id, created_at desc);
create index on follows (following_id, created_at desc);
