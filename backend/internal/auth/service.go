package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid credentials")
)

// Mailer はパスワード再設定の管理者承認フロー(README.md参照)で使うメール送信の抽象。
// 実装は internal/mail パッケージが提供する。
type Mailer interface {
	SendAdminApprovalRequest(ctx context.Context, requestID, userEmail string) error
	SendNewPassword(ctx context.Context, userEmail, newPassword string) error
}

type Service struct {
	repo             *Repository
	tokens           *TokenIssuer
	mailer           Mailer
	passwordResetTTL time.Duration
}

func NewService(repo *Repository, tokens *TokenIssuer, mailer Mailer, passwordResetTTL time.Duration) *Service {
	return &Service{repo: repo, tokens: tokens, mailer: mailer, passwordResetTTL: passwordResetTTL}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) Register(ctx context.Context, email, password, displayName string) (*User, error) {
	if _, err := s.repo.GetUserByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyRegistered
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateUser(ctx, email, hash, displayName)
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, *TokenPair, error) {
	u, err := s.repo.GetUserByEmail(ctx, email)
	if errors.Is(err, ErrNotFound) {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, err
	}
	if !checkPassword(u.PasswordHash, password) {
		return nil, nil, ErrInvalidCredentials
	}

	pair, err := s.issueTokenPair(u.ID)
	if err != nil {
		return nil, nil, err
	}
	return u, pair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	userID, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.issueTokenPair(userID)
}

func (s *Service) issueTokenPair(userID string) (*TokenPair, error) {
	access, err := s.tokens.IssueAccessToken(userID)
	if err != nil {
		return nil, err
	}
	refresh, err := s.tokens.IssueRefreshToken(userID)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *Service) Me(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// RequestPasswordReset はパスワード再設定を申請する。ユーザーの実在有無に関わらず
// 呼び出し元には同じレスポンスを返す想定(メールアドレス列挙攻撃を避けるため)。
// 対象ユーザーが実在する場合のみ、管理者宛の承認依頼メールを送信する。
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	u, err := s.repo.GetUserByEmail(ctx, email)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	req, err := s.repo.CreatePasswordResetRequest(ctx, u.ID, time.Now().Add(s.passwordResetTTL))
	if err != nil {
		return err
	}

	return s.mailer.SendAdminApprovalRequest(ctx, req.ID, u.Email)
}

// ApprovePasswordReset は管理者の空メール返信を検知したIMAPポーラー(internal/mail)から
// 呼び出される。新パスワードを自動生成してユーザーに送信する。
func (s *Service) ApprovePasswordReset(ctx context.Context, requestID string) error {
	req, err := s.repo.GetPendingPasswordResetRequest(ctx, requestID)
	if err != nil {
		return err
	}
	if time.Now().After(req.ExpiresAt) {
		return s.repo.ExpireStalePasswordResetRequests(ctx)
	}

	u, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	newPassword, err := generateRandomPassword()
	if err != nil {
		return err
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePasswordHash(ctx, u.ID, hash); err != nil {
		return err
	}
	if err := s.repo.ApprovePasswordResetRequest(ctx, req.ID); err != nil {
		return err
	}

	return s.mailer.SendNewPassword(ctx, u.Email, newPassword)
}
