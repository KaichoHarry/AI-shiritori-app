package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/gemini"
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/settings"
)

var (
	ErrInvalidMode      = errors.New("invalid mode")
	ErrSessionFinished  = errors.New("session already finished")
	ErrWrongModeForCall = errors.New("this endpoint does not apply to the session's mode")
)

type Service struct {
	repo         *Repository
	settingsRepo *settings.Repository
	gemini       *gemini.Client
}

func NewService(repo *Repository, settingsRepo *settings.Repository, geminiClient *gemini.Client) *Service {
	return &Service{repo: repo, settingsRepo: settingsRepo, gemini: geminiClient}
}

// CreateSession は新規セッションを作成する。difficulty/toneが指定されない場合は
// ユーザー設定のデフォルト値を使用する(DESIGN.md 7.3節)。
func (s *Service) CreateSession(ctx context.Context, userID string, mode Mode, difficulty, tone *string) (*Session, error) {
	switch mode {
	case ModeSolo, ModeVsAI, ModeFreeTalk:
	default:
		return nil, ErrInvalidMode
	}

	userSettings, err := s.settingsRepo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user settings: %w", err)
	}

	var difficultyPtr *string
	if mode == ModeVsAI {
		d := string(userSettings.Mode2Difficulty)
		if difficulty != nil && *difficulty != "" {
			d = *difficulty
		}
		difficultyPtr = &d
	}

	var tonePtr *string
	if mode == ModeFreeTalk {
		t := string(userSettings.Mode3Tone)
		if tone != nil && *tone != "" {
			t = *tone
		}
		tonePtr = &t
	}

	return s.repo.CreateSession(ctx, userID, mode, difficultyPtr, tonePtr)
}

func (s *Service) ListSessions(ctx context.Context, userID string, limit, offset int) ([]*Session, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListSessions(ctx, userID, limit, offset)
}

func (s *Service) GetSession(ctx context.Context, sessionID, userID string) (*Session, error) {
	return s.repo.GetSession(ctx, sessionID, userID)
}

func (s *Service) GetWords(ctx context.Context, sessionID string) ([]*Word, error) {
	return s.repo.ListWords(ctx, sessionID)
}

func (s *Service) GetMessages(ctx context.Context, sessionID string) ([]*MessageRecord, error) {
	return s.repo.ListMessages(ctx, sessionID)
}

// EndSession はモード3向けの手動終了(DESIGN.md 7.3節 POST /api/games/{id}/end)。
func (s *Service) EndSession(ctx context.Context, sessionID, userID string) (*Session, error) {
	session, err := s.repo.GetSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session.Status == StatusFinished {
		return session, nil
	}
	if err := s.repo.FinishSession(ctx, sessionID, ResultEndedByUser); err != nil {
		return nil, err
	}
	return s.repo.GetSession(ctx, sessionID, userID)
}
