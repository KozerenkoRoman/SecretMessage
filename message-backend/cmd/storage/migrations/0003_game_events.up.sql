CREATE TABLE game_events (
    id serial PRIMARY KEY,
    room_id text NOT NULL,
    turn_id integer NOT NULL,
    event_id integer NOT NULL, 
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    state_before jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Твій унікальний індекс залишається на місці
CREATE UNIQUE INDEX IF NOT EXISTS idx_game_events_room_event ON game_events(room_id, event_id);