package gemini

import (
	"context"
	"errors"
	"fmt"
	"log"
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

// withTimeout は1回の試行あたりのタイムアウトを適用する(DESIGN.md 5.3節)。
func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.timeout)
}

const (
	maxRetryAttempts = 3
	retryBackoffBase = 500 * time.Millisecond
)

// isTransientError は、リトライすれば成功する見込みがあるエラーかどうかを判定する。
// 実際にモード3を大量テストしたところ、Gemini API(gemini-3.5-flash)は503(高負荷)・
// 504(タイムアウト)・429(レート制限)を頻繁に返すことが分かったため、これらと
// クライアント側のタイムアウト(context.DeadlineExceeded)についてのみ再試行する。
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case 429, 503, 504:
			return true
		}
	}
	return false
}

// withRetry はGemini API呼び出しを、一時的なエラー(isTransientError参照)に対して
// 最大maxRetryAttempts回まで短い間隔を空けて再試行する(DESIGN.md 5.3節。
// 当初は「リトライなし」の暫定案だったが、実測でモード3の応答生成が高い頻度で
// 一時的なエラーに遭遇することが分かったため、ユーザー確認の上リトライを追加した)。
func withRetry[T any](ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	var zero T
	var lastErr error
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * retryBackoffBase
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(backoff):
			}
			log.Printf("gemini: retrying after transient error (attempt %d/%d): %v", attempt+1, maxRetryAttempts, lastErr)
		}
		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isTransientError(err) {
			return zero, err
		}
	}
	return zero, lastErr
}
