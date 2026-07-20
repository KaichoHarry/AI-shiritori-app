package gemini

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

type Client struct {
	genai   *genai.Client
	model   string
	timeout time.Duration
}

// New はGemini APIクライアントを初期化する。APIキーはバックエンドの環境変数でのみ
// 管理し、フロントエンドには一切露出させない(DESIGN.md 5.3節)。
func New(ctx context.Context, apiKey, model string, timeout time.Duration) (*Client, error) {
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &Client{genai: c, model: model, timeout: timeout}, nil
}

// withTimeout はDESIGN.md 5.3節の「5秒タイムアウト・リトライなし」方針を適用する。
func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.timeout)
}
