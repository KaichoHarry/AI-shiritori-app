package settings

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Get(ctx context.Context, userID string) (*Settings, error) {
	var s Settings
	err := r.pool.QueryRow(ctx, `
		SELECT user_id::text, mode2_difficulty, mode3_tone, updated_at
		FROM user_settings WHERE user_id = $1
	`, userID).Scan(&s.UserID, &s.Mode2Difficulty, &s.Mode3Tone, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) Update(ctx context.Context, userID string, difficulty Difficulty, tone Tone) (*Settings, error) {
	var s Settings
	err := r.pool.QueryRow(ctx, `
		UPDATE user_settings
		SET mode2_difficulty = $2, mode3_tone = $3, updated_at = now()
		WHERE user_id = $1
		RETURNING user_id::text, mode2_difficulty, mode3_tone, updated_at
	`, userID, difficulty, tone).Scan(&s.UserID, &s.Mode2Difficulty, &s.Mode3Tone, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
