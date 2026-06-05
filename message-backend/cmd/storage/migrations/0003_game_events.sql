CREATE TABLE game_turns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id VARCHAR(255) NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    sequence_id INT NOT NULL, 
    action_type VARCHAR(50) NOT NULL, -- "APPLY" або "CHANCELLOR_RESOLVE"
    player_id VARCHAR(255) NOT NULL, 
    hand_index INT NOT NULL, -- для ChancellorResolve це буде KeepHandIndex
    target_id VARCHAR(255), 
    guess_card INT, -- тепер INT, повністю відповідає CardType
    played_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_room_sequence UNIQUE (room_id, sequence_id)
);

CREATE TABLE game_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id VARCHAR(255) NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    turn_id UUID REFERENCES game_turns(id) ON DELETE CASCADE,
    event_id BIGINT NOT NULL, 
    event_type VARCHAR(100) NOT NULL, 
    payload JSONB NOT NULL, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_game_turns_room ON game_turns(room_id);
CREATE INDEX idx_game_events_room ON game_events(room_id);