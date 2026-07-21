package gemini

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// Tone はモード3(文章しりとり)の会話トーン。DESIGN.md 6.2節のuser_settings.mode3_toneと対応する。
type Tone string

const (
	ToneFriendly Tone = "friendly"
	TonePolite   Tone = "polite"
	ToneComedy   Tone = "comedy"
)

// Speaker はモード3の会話履歴の発言者。
type Speaker string

const (
	SpeakerUser Speaker = "user"
	SpeakerAI   Speaker = "ai"
)

// Message はモード3の会話履歴1件分(DESIGN.md 6.2節のgame_messagesに対応)。
type Message struct {
	Speaker Speaker
	Content string
}

// GenerateReply はモード3(文章しりとり・フリートーク)のAI応答をGemini APIに生成させる
// (DESIGN.md 5.2節)。history は直近N件を呼び出し元(internal/game)が渡す想定。
//
// 実際に大量にテストしたところ、Gemini API(gemini-3.5-flash)が503(高負荷)・
// 504(タイムアウト)を高い頻度で返すことが分かったため、一時的なエラーには
// internal/gemini.withRetryで再試行する(DESIGN.md 5.3節、ユーザー確認済み)。
func (c *Client) GenerateReply(ctx context.Context, tone Tone, history []Message) (string, error) {
	temperature := float32(0.9)
	systemInstruction := genai.NewContentFromText(systemInstructionForTone(tone), genai.RoleUser)
	contents := toContents(history)

	reply, err := withRetry(ctx, func(attemptCtx context.Context) (string, error) {
		attemptCtx, cancel := c.withTimeout(attemptCtx)
		defer cancel()

		resp, err := c.genai.Models.GenerateContent(attemptCtx, c.model, contents, &genai.GenerateContentConfig{
			Temperature:       &temperature,
			SystemInstruction: systemInstruction,
		})
		if err != nil {
			return "", fmt.Errorf("generate reply: %w", err)
		}
		text := strings.TrimSpace(resp.Text())
		if text == "" {
			return "", fmt.Errorf("gemini returned an empty reply")
		}
		return text, nil
	})
	if err != nil {
		return "", err
	}
	return reply, nil
}

func toContents(history []Message) []*genai.Content {
	contents := make([]*genai.Content, 0, len(history))
	for _, m := range history {
		var role genai.Role = genai.RoleUser
		if m.Speaker == SpeakerAI {
			role = genai.RoleModel
		}
		contents = append(contents, genai.NewContentFromText(m.Content, role))
	}
	return contents
}

func systemInstructionForTone(tone Tone) string {
	base := "あなたは「文章しりとり」という言葉遊びをしながら日常会話を続けるAIです。\n" +
		"文章しりとりのルール: 直前の発言(あなた自身の発言でも相手の発言でも構いません)の文末の音から、" +
		"次の発言を始めてください。単語単位の厳密なしりとりではなく、自然な文章のつながりを大事にしてください。\n" +
		"守ってほしいこと:\n" +
		"- 発言は1〜2文程度の短い返答にしてください。\n" +
		"- 「ん」の音で終わる発言は、その時点で会話が終了する合図になるため、できるだけ避けてください。\n" +
		"- どのような内容の発言に対しても、必ず何らかの返答をしてください。返答を拒否したり、空の返答を返したりしないでください。\n\n"

	switch tone {
	case TonePolite:
		return base + "口調は丁寧語・敬語にしてください。落ち着いた大人な雰囲気で会話してください。"
	case ToneComedy:
		return base + "お笑い芸人のようなノリで、コント風の大げさなツッコミやボケを交えて会話してください。"
	default:
		return base + "口調はフレンドリーでカジュアルにしてください。親しい友人と話すような雰囲気にしてください。"
	}
}
