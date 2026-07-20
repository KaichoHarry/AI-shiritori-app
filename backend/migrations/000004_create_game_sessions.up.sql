CREATE TABLE game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mode TEXT NOT NULL CHECK (mode IN ('solo', 'vs_ai', 'free_talk')),
    difficulty TEXT CHECK (difficulty IN ('easy', 'normal', 'hard')),
    tone TEXT CHECK (tone IN ('friendly', 'polite', 'comedy')),
    status TEXT NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'finished')),
    result TEXT CHECK (result IN ('win', 'lose', 'ended_by_n', 'ended_by_duplicate', 'ended_by_user')),
    turn_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ
);

CREATE INDEX idx_game_sessions_user_id ON game_sessions(user_id);
