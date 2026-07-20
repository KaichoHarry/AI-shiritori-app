CREATE TABLE game_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    seq_no INTEGER NOT NULL,
    speaker TEXT NOT NULL CHECK (speaker IN ('user', 'ai')),
    content TEXT NOT NULL,
    ending_sound TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (session_id, seq_no)
);

CREATE INDEX idx_game_messages_session_id ON game_messages(session_id);
