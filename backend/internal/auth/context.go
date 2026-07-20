package auth

import "context"

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func userIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userIDContextKey).(string)
	return v
}

// UserIDFromContext は他パッケージ(internal/game等)がリクエストコンテキストから
// 認証済みユーザーIDを取り出すための公開関数。
func UserIDFromContext(ctx context.Context) string {
	return userIDFromContext(ctx)
}
