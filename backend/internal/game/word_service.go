package game

import (
	"context"
	"math/rand"

	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/shiritori"
)

// Reason はDESIGN.md 7.4節のreason値。
//
// AI側の単語は辞書(internal/shiritori.RandomWord)から選ぶため接続・重複・「ん」終端
// といった不正な語を返すことは無く、AI敗北の理由は「その頭文字の候補が辞書から尽きた」
// (ai_word_not_found)か「わざと失敗する確率に当たった」(ai_generation_failed)の
// いずれかのみになる。
type Reason string

const (
	ReasonNone               Reason = ""
	ReasonWordNotFound       Reason = "word_not_found"
	ReasonEndsWithN          Reason = "ends_with_n"
	ReasonConnectionMismatch Reason = "connection_mismatch"
	ReasonDuplicateWord      Reason = "duplicate_word"
	ReasonAIWordNotFound     Reason = "ai_word_not_found"
	ReasonAIGenerationFailed Reason = "ai_generation_failed"
)

// モード2のAI側は、Gemini APIを使わず辞書(internal/shiritori.RandomWord)からの
// ランダム選択に決定的に切り替えている。理由: Gemini経由だとAIが接続に失敗しやすく
// しりとりがすぐ終わってしまい、ユーザーからより長く続くようにしたいと要望された。
// 難易度ごとの挙動(ユーザー確認済み):
//   - hard:   常に辞書からその頭文字・「ん」以外で終わる語を選び続ける
//     (その頭文字の候補が尽きたときのみAI側の負けとして終了)
//   - normal: 49ラリーまではhardと同じ。50ラリー目以降は毎ターン1%の確率でわざと負ける
//   - easy:   毎ターン1%の確率でわざと負ける(hard/normalと同じ辞書選択に加えて)
const (
	aiIntentionalFailureRate           = 0.01
	aiNormalDifficultyGracePeriodRally = 49
)

func shouldAIIntentionallyFail(difficulty string, aiRally int) bool {
	switch difficulty {
	case "easy":
		return rand.Float64() < aiIntentionalFailureRate
	case "normal":
		if aiRally > aiNormalDifficultyGracePeriodRally {
			return rand.Float64() < aiIntentionalFailureRate
		}
		return false
	default: // hard、または未指定
		return false
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

	aiWordsSoFar, err := s.repo.CountWordsBySpeaker(ctx, session.ID, SpeakerAI)
	if err != nil {
		return nil, err
	}
	aiRally := aiWordsSoFar + 1 // これから出す単語のラリー番号(1始まり)

	if shouldAIIntentionallyFail(difficulty, aiRally) {
		return s.finishAsPlayerWin(ctx, session.ID, ReasonAIGenerationFailed, playerWordView)
	}

	aiReading, found := shiritori.RandomWord(shiritori.LastSound(playerReading), usedReadings)
	if !found {
		// その頭文字から始まり「ん」で終わらない未使用語が辞書に無くなった場合、AIの負け。
		return s.finishAsPlayerWin(ctx, session.ID, ReasonAIWordNotFound, playerWordView)
	}

	aiSeq := playerSeq + 1
	if err := s.repo.AddWord(ctx, session.ID, aiSeq, SpeakerAI, aiReading, aiReading); err != nil {
		return nil, err
	}

	return &WordResult{
		Accepted:   true,
		PlayerWord: playerWordView,
		AIWord:     &WordView{Word: aiReading, Reading: aiReading},
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

func resultPtr(r Result) *Result {
	return &r
}
