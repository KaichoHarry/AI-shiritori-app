package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// Difficulty はモード2(AI対戦)の難易度。DESIGN.md 6.2節のuser_settings.mode2_difficultyと対応する。
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyNormal Difficulty = "normal"
	DifficultyHard   Difficulty = "hard"
)

type wordSuggestion struct {
	Word string `json:"word"`
}

// GenerateWord はモード2(AI対戦)のAI側の単語をGemini APIに生成させる(DESIGN.md 5.1節)。
// 応答の正当性(お尻の文字と繋がっているか等)の検証は呼び出し元(internal/shiritori.Judge等)
// が行う。わざと失敗させる確率的制御はコスト削減のため、この関数を呼ぶ前に呼び出し元が
// 乱数で判定する想定(DESIGN.md 5.1節)。
func (c *Client) GenerateWord(ctx context.Context, requiredFirstSound string, usedWords []string, difficulty Difficulty) (string, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	temperature := temperatureForDifficulty(difficulty)
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"word": {
				Type:        genai.TypeString,
				Description: "しりとりの次の単語(ひらがな・カタカナ・漢字のいずれかで1語)",
			},
		},
		Required: []string{"word"},
	}

	resp, err := c.genai.Models.GenerateContent(ctx, c.model, genai.Text(buildWordPrompt(requiredFirstSound, usedWords, difficulty)), &genai.GenerateContentConfig{
		Temperature:      &temperature,
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	})
	if err != nil {
		return "", fmt.Errorf("generate word: %w", err)
	}

	var suggestion wordSuggestion
	if err := json.Unmarshal([]byte(resp.Text()), &suggestion); err != nil {
		return "", fmt.Errorf("parse word suggestion: %w", err)
	}
	if strings.TrimSpace(suggestion.Word) == "" {
		return "", fmt.Errorf("gemini returned an empty word")
	}
	return suggestion.Word, nil
}

func buildWordPrompt(requiredFirstSound string, usedWords []string, difficulty Difficulty) string {
	var b strings.Builder
	b.WriteString("あなたはしりとりゲームのプレイヤーです。次のルールに従って、しりとりの単語を1つだけ考えてください。\n\n")
	fmt.Fprintf(&b, "- 次に出す単語は「%s」から始まる単語でなければなりません。\n", requiredFirstSound)
	b.WriteString("- 「ん」で終わる単語は絶対に選ばないでください。\n")
	if len(usedWords) > 0 {
		fmt.Fprintf(&b, "- 以下の単語は既に使用済みなので、選ばないでください: %s\n", strings.Join(usedWords, "、"))
	}
	b.WriteString(difficultyInstruction(difficulty))
	b.WriteString("\n実在する日本語の単語を1つだけ選び、指定された形式で回答してください。")
	return b.String()
}

func difficultyInstruction(d Difficulty) string {
	switch d {
	case DifficultyEasy:
		return "- 難易度は「易しい」なので、小学生でも知っているような簡単で短い単語を選んでください。\n"
	case DifficultyHard:
		return "- 難易度は「難しい」なので、難読語や長い単語、専門用語などを積極的に選んでください。\n"
	default:
		return "- 難易度は「普通」なので、一般的な語彙の中から単語を選んでください。\n"
	}
}

func temperatureForDifficulty(d Difficulty) float32 {
	switch d {
	case DifficultyEasy:
		return 0.3
	case DifficultyHard:
		return 1.0
	default:
		return 0.7
	}
}
