package sentence

import (
	"strings"

	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/shiritori"
)

// trailingPunctuation は文末の音を抽出する前に取り除く句読点・記号(DESIGN.md 8.1節1.参照)。
const trailingPunctuation = "。.!?!?、,　 \n\r\t…「」『』\"'"

// ExtractEndingSound は発言(文章)から、末尾の句読点・記号を除去したうえで
// 文末の音(モーラ)を抽出する。しりとり判定と同じかな正規化・モーラ分割ロジックを流用する。
func ExtractEndingSound(content string) string {
	trimmed := strings.TrimRight(content, trailingPunctuation)
	if trimmed == "" {
		return ""
	}
	reading := shiritori.Analyze(trimmed).Reading
	return shiritori.LastSound(reading)
}

// EndReason はチェーン終了理由。DESIGN.md 2.3節のresult値と対応する。
type EndReason string

const (
	EndReasonNone      EndReason = ""
	EndReasonN         EndReason = "ended_by_n"
	EndReasonDuplicate EndReason = "ended_by_duplicate"
)

// Result はEvaluateの評価結果。
type Result struct {
	EndingSound string
	Ended       bool
	Reason      EndReason
}

// Evaluate はDESIGN.md 8.1節の3.〜5.を実装する。接続チェック(2.)は致命的エラーに
// せず会話として受け流す設計方針のため、ここでは判定しない(AI応答生成側のプロンプトで
// 吸収する。5.2節参照)。
func Evaluate(content string, previousContents []string) Result {
	endingSound := ExtractEndingSound(content)

	if endingSound == "ん" {
		return Result{EndingSound: endingSound, Ended: true, Reason: EndReasonN}
	}

	for _, prev := range previousContents {
		if prev == content {
			return Result{EndingSound: endingSound, Ended: true, Reason: EndReasonDuplicate}
		}
	}

	return Result{EndingSound: endingSound, Ended: false, Reason: EndReasonNone}
}
