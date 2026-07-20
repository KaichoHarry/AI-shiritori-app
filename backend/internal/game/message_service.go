package game

import (
	"context"
	"log"

	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/gemini"
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/sentence"
)

const recentHistoryLimit = 20 // 直近10往復(DESIGN.md 5.2節の例)= 20発言

type MessageView struct {
	Content     string
	EndingSound string
}

type MessageResult struct {
	UserMessage *MessageView
	AIMessage   *MessageView
	Status      Status
	Result      *Result
}

// SubmitMessage はモード3(文章しりとり・フリートーク)の発言送信処理
// (DESIGN.md 2.3・5.2・7.5・8.1章)。
func (s *Service) SubmitMessage(ctx context.Context, sessionID, userID, content string) (*MessageResult, error) {
	session, err := s.repo.GetSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session.Mode != ModeFreeTalk {
		return nil, ErrWrongModeForCall
	}
	if session.Status == StatusFinished {
		return nil, ErrSessionFinished
	}

	previousContents, err := s.repo.AllContents(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	userEval := sentence.Evaluate(content, previousContents)
	userView := &MessageView{Content: content, EndingSound: userEval.EndingSound}

	if userEval.Ended {
		nextSeq := session.TurnCount + 1
		if err := s.repo.AddMessage(ctx, sessionID, nextSeq, SpeakerUser, content, userEval.EndingSound); err != nil {
			return nil, err
		}
		result := mapEndReason(userEval.Reason)
		if err := s.repo.FinishSession(ctx, sessionID, result); err != nil {
			return nil, err
		}
		return &MessageResult{UserMessage: userView, Status: StatusFinished, Result: resultPtr(result)}, nil
	}

	// AIの返答を生成する前に外部API呼び出しを行い、失敗時は何もDBに書き込まずエラーを返す
	// (ユーザーが安全に再送信できるようにするため)。
	recent, err := s.repo.RecentMessages(ctx, sessionID, recentHistoryLimit)
	if err != nil {
		return nil, err
	}
	history := make([]gemini.Message, 0, len(recent)+1)
	for _, m := range recent {
		history = append(history, gemini.Message{Speaker: toGeminiSpeaker(m.Speaker), Content: m.Content})
	}
	history = append(history, gemini.Message{Speaker: gemini.SpeakerUser, Content: content})

	tone := string(gemini.ToneFriendly)
	if session.Tone != nil {
		tone = *session.Tone
	}

	aiReply, err := s.gemini.GenerateReply(ctx, gemini.Tone(tone), history)
	if err != nil {
		log.Printf("mode3: gemini reply generation failed: %v", err)
		return nil, err
	}

	nextSeq := session.TurnCount + 1
	if err := s.repo.AddMessage(ctx, sessionID, nextSeq, SpeakerUser, content, userEval.EndingSound); err != nil {
		return nil, err
	}

	aiEval := sentence.Evaluate(aiReply, append(previousContents, content))
	aiView := &MessageView{Content: aiReply, EndingSound: aiEval.EndingSound}

	aiSeq := nextSeq + 1
	if err := s.repo.AddMessage(ctx, sessionID, aiSeq, SpeakerAI, aiReply, aiEval.EndingSound); err != nil {
		return nil, err
	}

	if aiEval.Ended {
		result := mapEndReason(aiEval.Reason)
		if err := s.repo.FinishSession(ctx, sessionID, result); err != nil {
			return nil, err
		}
		return &MessageResult{UserMessage: userView, AIMessage: aiView, Status: StatusFinished, Result: resultPtr(result)}, nil
	}

	return &MessageResult{UserMessage: userView, AIMessage: aiView, Status: StatusInProgress}, nil
}

func toGeminiSpeaker(sp Speaker) gemini.Speaker {
	if sp == SpeakerAI {
		return gemini.SpeakerAI
	}
	return gemini.SpeakerUser
}

func mapEndReason(r sentence.EndReason) Result {
	if r == sentence.EndReasonDuplicate {
		return ResultEndedByDuplicate
	}
	return ResultEndedByN
}
