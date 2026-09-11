create table if not exists user_stats (
    user_id uuid primary key references users(id) on delete cascade,
    games_played int not null default 0,
    games_won int not null default 0,
    rounds_played int not null default 0,
    rounds_won int not null default 0,
    spy_bonuses int not null default 0,
    total_score int not null default 0, 
    updated_at timestamptz not null default now()
);

create index if not exists idx_user_stats_score on user_stats(total_score desc, games_won desc);