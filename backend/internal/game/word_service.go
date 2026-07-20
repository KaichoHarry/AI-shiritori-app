package game

import (
	"context"
	"errors"
	"math/rand"

	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/gemini"
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/shiritori"
)

// Reason はDESIGN.md 7.4節のreason値。AI側の理由(ai_*)は明示的な一覧が
// word_not_found/timeout/generation_failedしか無かったため、プレイヤー側の
// 判定理由と対応させたai_ends_with_n等をClaude Codeの判断で追加している。
type Reason string

const (
	ReasonNone                 Reason = ""
	ReasonWordNotFound         Reason = "word_not_found"
	ReasonEndsWithN            Reason = "ends_with_n"
	ReasonConnectionMismatch   Reason = "connection_mismatch"
	ReasonDuplicateWord        Reason = "duplicate_word"
	ReasonAIWordNotFound       Reason = "ai_word_not_found"
	ReasonAIEndsWithN          Reason = "ai_ends_with_n"
	ReasonAIConnectionMismatch Reason = "ai_connection_mismatch"
	ReasonAIDuplicateWord      Reason = "ai_duplicate_word"
	ReasonAITimeout            Reason = "ai_timeout"
	ReasonAIGenerationFailed   Reason = "ai_generation_failed"
)

// intentionalFailureProbability はDESIGN.md 5.1節「AIがわざと失敗する確率」の暫定値。
// 易しい=20%は設計書の例をそのまま採用し、普通・難しいはClaude Codeが妥当な値を暫定的に設定した。
func intentionalFailureProbability(difficulty string) float64 {
	switch difficulty {
	case string(gemini.DifficultyEasy):
		return 0.2
	case string(gemini.DifficultyHard):
		return 0.02
	default:
		return 0.08
	}
}

type WordView struct {
	Word    string
	Reading string
}

type WordResult struct {
	Accepted   bool
	Fatal      bool
	Reason     Reason
	Message    string
	PlayerWord *WordView
	AIWord     *WordView
	Status     Status
	Result     *Result
}

// SubmitWord はモード1(ソロプレイ)・モード2(AI対戦)共通のプレイヤー単語送信処理
// (DESIGN.md 2.1・2.2・7.4・8章)。
func (s *Service) SubmitWord(ctx context.Context, sessionID, userID, rawWord string) (*WordResult, error) {
	session, err := s.repo.GetSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session.Mode != ModeSolo && session.Mode != ModeVsAI {
		return nil, ErrWrongModeForCall
	}
	if session.Status == StatusFinished {
		return nil, ErrSessionFinished
	}

	usedReadings, err := s.repo.UsedReadings(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	previousReading, err := s.repo.LastWordReading(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	playerJudge := shiritori.Judge(rawWord, previousReading, usedReadings)

	if !playerJudge.Accepted && !playerJudge.Fatal {
		return &WordResult{
			Accepted: false,
			Fatal:    false,
			Reason:   Reason(playerJudge.Reason),
			Message:  nonFatalMessage(playerJudge.Reason, rawWord),
			Status:   session.Status,
		}, nil
	}

	nextSeq := session.TurnCount + 1
	if err := s.repo.AddWord(ctx, sessionID, nextSeq, SpeakerUser, rawWord, playerJudge.Reading); err != nil {
		return nil, err
	}

	if !playerJudge.Accepted {
		// 致命的エラー(ends_with_n / connection_mismatch / duplicate_word): プレイヤーの負け。
		if err := s.repo.FinishSession(ctx, sessionID, ResultLose); err != nil {
			return nil, err
		}
		return &WordResult{
			Accepted:   false,
			Fatal:      true,
			Reason:     Reason(playerJudge.Reason),
			PlayerWord: &WordView{Word: rawWord, Reading: playerJudge.Reading},
			Status:     StatusFinished,
			Result:     resultPtr(ResultLose),
		}, nil
	}

	playerWordView := &WordView{Word: rawWord, Reading: playerJudge.Reading}

	if session.Mode == ModeSolo {
		return &WordResult{
			Accepted:   true,
			PlayerWord: playerWordView,
			Status:     StatusInProgress,
		}, nil
	}

	// モード2: AI側のターン。
	usedReadings[playerJudge.Reading] = true
	return s.playAITurn(ctx, session, nextSeq, playerJudge.Reading, usedReadings, playerWordView)
}

func (s *Service) playAITurn(ctx context.Context, session *Session, playerSeq int, playerReading string, usedReadings map[string]bool, playerWordView *WordView) (*WordResult, error) {
	difficulty := ""
	if session.Difficulty != nil {
		difficulty = *session.Difficulty
	}

	if rand.Float64() < intentionalFailureProbability(difficulty) {
		return s.finishAsPlayerWin(ctx, session.ID, ReasonAIGenerationFailed, playerWordView)
	}

	usedWords := make([]string, 0, len(usedReadings))
	for reading := range usedReadings {
		usedWords = append(usedWords, reading)
	}

	aiWord, err := s.gemini.GenerateWord(ctx, shiritori.LastSound(playerReading), usedWords, gemini.Difficulty(difficulty))
	if err != nil {
		reason := ReasonAIGenerationFailed
		if errors.Is(err, context.DeadlineExceeded) {
			reason = ReasonAITimeout
		}
		return s.finishAsPlayerWin(ctx, session.ID, reason, playerWordView)
	}

	aiJudge := shiritori.Judge(aiWord, playerReading, usedReadings)
	if !aiJudge.Accepted {
		return s.finishAsPlayerWin(ctx, session.ID, mapAIReason(aiJudge.Reason), playerWordView)
	}

	aiSeq := playerSeq + 1
	if err := s.repo.AddWord(ctx, session.ID, aiSeq, SpeakerAI, aiWord, aiJudge.Reading); err != nil {
		return nil, err
	}

	return &WordResult{
		Accepted:   true,
		PlayerWord: playerWordView,
		AIWord:     &WordView{Word: aiWord, Reading: aiJudge.Reading},
		Status:     StatusInProgress,
	}, nil
}

func (s *Service) finishAsPlayerWin(ctx context.Context, sessionID string, reason Reason, playerWordView *WordView) (*WordResult, error) {
	if err := s.repo.FinishSession(ctx, sessionID, ResultWin); err != nil {
		return nil, err
	}
	return &WordResult{
		Accepted:   true,
		Fatal:      true,
		Reason:     reason,
		PlayerWord: playerWordView,
		Status:     StatusFinished,
		Result:     resultPtr(ResultWin),
	}, nil
}

func nonFatalMessage(reason shiritori.Reason, rawWord string) string {
	switch reason {
	case shiritori.ReasonInvalidCharacters:
		return "ひらがな・カタカナのみで入力してください。"
	default:
		return "「" + rawWord + "」という言葉は辞書に見つかりませんでした。別の言葉を入力してください。"
	}
}

func mapAIReason(r shiritori.Reason) Reason {
	switch r {
	case shiritori.ReasonWordNotFound:
		return ReasonAIWordNotFound
	case shiritori.ReasonEndsWithN:
		return ReasonAIEndsWithN
	case shiritori.ReasonConnectionMismatch:
		return ReasonAIConnectionMismatch
	case shiritori.ReasonDuplicateWord:
		return ReasonAIDuplicateWord
	default:
		return ReasonAIGenerationFailed
	}
}

func resultPtr(r Result) *Result {
	return &r
}
