CREATE TABLE game_words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    seq_no INTEGER NOT NULL,
    speaker TEXT NOT NULL CHECK (speaker IN ('user', 'ai')),
    word TEXT NOT NULL,
    reading TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (session_id, seq_no)
);

CREATE INDEX idx_game_words_session_id ON game_words(session_id);
