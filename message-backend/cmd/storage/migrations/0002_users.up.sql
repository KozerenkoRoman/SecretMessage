create table if not exists users (
    id uuid primary key default gen_random_uuid(),
    username varchar(50) unique not null,
    email varchar(255) unique not null,
    password_hash varchar(255) not null,
    user_role varchar(20) not null default 'user', 
    is_banned boolean not null default false,
    ban_reason text,
    banned_at timestamptz,
    avatar_seed varchar(255) not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_users_username on users(username);
create index if not exists idx_users_email on users(email);