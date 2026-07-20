package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, displayName string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id::text, email, password_hash, display_name, created_at, updated_at
	`, email, passwordHash, displayName).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO user_settings (user_id) VALUES ($1)
	`, u.ID)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, email, password_hash, display_name, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, email, password_hash, display_name, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2
	`, passwordHash, userID)
	return err
}

func (r *Repository) CreatePasswordResetRequest(ctx context.Context, userID string, expiresAt time.Time) (*PasswordResetRequest, error) {
	var req PasswordResetRequest
	err := r.pool.QueryRow(ctx, `
		INSERT INTO password_reset_requests (user_id, expires_at)
		VALUES ($1, $2)
		RETURNING id::text, user_id::text, status, admin_notified_at, approved_at, expires_at, created_at
	`, userID, expiresAt).Scan(
		&req.ID, &req.UserID, &req.Status, &req.AdminNotifiedAt, &req.ApprovedAt, &req.ExpiresAt, &req.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Repository) GetPendingPasswordResetRequest(ctx context.Context, requestID string) (*PasswordResetRequest, error) {
	var req PasswordResetRequest
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, status, admin_notified_at, approved_at, expires_at, created_at
		FROM password_reset_requests
		WHERE id = $1 AND status = 'pending'
	`, requestID).Scan(
		&req.ID, &req.UserID, &req.Status, &req.AdminNotifiedAt, &req.ApprovedAt, &req.ExpiresAt, &req.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Repository) ApprovePasswordResetRequest(ctx context.Context, requestID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE password_reset_requests SET status = 'approved', approved_at = now()
		WHERE id = $1
	`, requestID)
	return err
}

func (r *Repository) ExpireStalePasswordResetRequests(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE password_reset_requests SET status = 'expired'
		WHERE status = 'pending' AND expires_at < now()
	`)
	return err
}
