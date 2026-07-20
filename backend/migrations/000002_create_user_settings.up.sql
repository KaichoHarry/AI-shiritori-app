CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    mode2_difficulty TEXT NOT NULL DEFAULT 'normal' CHECK (mode2_difficulty IN ('easy', 'normal', 'hard')),
    mode3_tone TEXT NOT NULL DEFAULT 'friendly' CHECK (mode3_tone IN ('friendly', 'polite', 'comedy')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
