create table rooms (
    id varchar(255) primary key,
    state jsonb not null,
    updated_at timestamptz not null default now()
);

create index idx_rooms_updated_at on rooms(updated_at);