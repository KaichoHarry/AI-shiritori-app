package shiritori

import (
	"strings"
)

// AnalyzeResult は入力語の辞書存在チェック結果。
type AnalyzeResult struct {
	// Reading は入力語をひらがなに正規化した読み。
	Reading string
	// Known は辞書引き上、実在する単語(名詞)として認識できるかどうか。
	Known bool
}

// Analyze は入力語(ひらがな・カタカナのみを想定。呼び出し元のJudgeがIsKanaOnlyで
// 事前に弾く)をひらがなに正規化し、IPADIC辞書の名詞の読みとして実在するかどうかを
// 判定する(DESIGN.md 4.1・4.2・8章参照)。
//
// 以前はkagomeの形態素解析器でこの文字列を再トークン化し、既知語として認識できるかを
// 見る方式だったが、以下の不整合が判明したため、読み(ひらがな)の集合を直接引く方式
// (internal/shiritori.IsKnownReading、AIの単語バンクと共通のデータソース)に変更した。
//   - 「花火」はトークナイザでKnown=trueになるが、読みの「はなび」(ひらがな)を
//     そのままトークン化すると未知語判定になってしまう(漢字の見出し語をひらがなの
//     読みだけで再認識できないケースがある)。
//   - ひらがな⇔カタカナ変換フォールバックの副作用で、「ぁ」「っ」のような
//     小さい仮名1文字が、IPADIC辞書内の記号的なエントリを拾って誤って既知語判定
//     されていた。
func Analyze(word string) AnalyzeResult {
	hiragana := katakanaToHiragana(word)
	return AnalyzeResult{
		Reading: hiragana,
		Known:   IsKnownReading(hiragana),
	}
}

// katakanaToHiragana はカタカナをひらがなに変換する。カタカナ範囲外の文字はそのまま残す。
func katakanaToHiragana(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 0x30A1 && r <= 0x30F6 {
			r -= 0x60
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ToHiragana はkatakanaToHiraganaの公開版(internal/sentence等の他パッケージから利用する)。
func ToHiragana(s string) string {
	return katakanaToHiragana(s)
}

// ToKatakana はひらがなをカタカナに変換する(次に入力すべき音のヒント表示で、
// ひらがな・カタカナ両方の表記を示すために使う)。ひらがな範囲外の文字はそのまま残す。
func ToKatakana(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 0x3041 && r <= 0x3096 {
			r += 0x60
		}
		b.WriteRune(r)
	}
	return b.String()
}

// IsHiragana は文字がひらがな(長音「ー」を含む)かどうかを判定する。
func IsHiragana(r rune) bool {
	return (r >= 0x3041 && r <= 0x3096) || r == longVowelMark
}

// IsKatakana は文字がカタカナ(長音「ー」を含む)かどうかを判定する。
func IsKatakana(r rune) bool {
	return (r >= 0x30A1 && r <= 0x30FC)
}

// IsKanaOnly は語がひらがな・カタカナのみで構成されているかどうかを判定する
// (モード1・2の入力制約。漢字・英数字・記号等は不可)。
func IsKanaOnly(word string) bool {
	if word == "" {
		return false
	}
	for _, r := range word {
		if !IsHiragana(r) && !IsKatakana(r) {
			return false
		}
	}
	return true
}
