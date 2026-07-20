package sentence

import (
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/shiritori"
)

// ExtractEndingSound は発言(文章)の末尾から遡り、最初に見つかった「有効な文字」
// (ひらがな・カタカナ・半角英字・全角英字)をもとに文末の音を抽出する。
//
// 句読点・記号・絵文字等の「文末に付け足された飾り」を一律に取り除く方式ではなく、
// 有効な文字が見つかるまで遡る方式を採用している(ユーザー確認済み)。この文字種の
// 制約はフロントエンド側でもユーザーに周知する想定。
func ExtractEndingSound(content string) string {
	runes := []rune(content)
	idx := -1
	for i := len(runes) - 1; i >= 0; i-- {
		if isAllowedEndingRune(runes[i]) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ""
	}

	if isAlphabet(runes[idx]) {
		return string(runes[idx])
	}

	// ひらがな・カタカナの場合は、見つかった文字までの接頭辞をひらがなに正規化したうえで
	// しりとり判定モジュールのモーラ分割ロジック(長音・拗音の扱い)を流用する。
	prefix := shiritori.ToHiragana(string(runes[:idx+1]))
	return shiritori.LastSound(prefix)
}

func isAllowedEndingRune(r rune) bool {
	return shiritori.IsHiragana(r) || shiritori.IsKatakana(r) || isAlphabet(r)
}

func isAlphabet(r rune) bool {
	switch {
	case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z': // 半角英字
		return true
	case r >= 0xFF21 && r <= 0xFF3A, r >= 0xFF41 && r <= 0xFF5A: // 全角英字
		return true
	default:
		return false
	}
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
