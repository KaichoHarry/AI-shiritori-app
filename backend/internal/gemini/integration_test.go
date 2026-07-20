package gemini

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// これらのテストは実際にGemini APIを呼び出す結合テスト。GEMINI_API_KEYが
// 設定されていない場合はスキップする(通常のユニットテストでは実行されない)。
func testClient(t *testing.T) *Client {
	t.Helper()
	_ = godotenv.Load("../../../.env", "../../.env")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set; skipping Gemini integration test")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-3.5-flash"
	}

	c, err := New(context.Background(), apiKey, model, 10*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestGenerateWordIntegration(t *testing.T) {
	c := testClient(t)

	word, err := c.GenerateWord(context.Background(), "り", []string{"りんご"}, DifficultyNormal)
	if err != nil {
		t.Fatalf("GenerateWord: %v", err)
	}
	if word == "" {
		t.Fatal("GenerateWord returned empty word")
	}
	t.Logf("generated word: %s", word)
}

func TestGenerateReplyIntegration(t *testing.T) {
	c := testClient(t)

	reply, err := c.GenerateReply(context.Background(), ToneFriendly, []Message{
		{Speaker: SpeakerUser, Content: "今日は天気がいいね"},
	})
	if err != nil {
		t.Fatalf("GenerateReply: %v", err)
	}
	if reply == "" {
		t.Fatal("GenerateReply returned empty reply")
	}
	t.Logf("generated reply: %s", reply)
}
