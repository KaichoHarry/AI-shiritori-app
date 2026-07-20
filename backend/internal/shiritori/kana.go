package shiritori

import (
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

var sharedTokenizer *tokenizer.Tokenizer

func init() {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic("shiritori: failed to initialize tokenizer: " + err.Error())
	}
	sharedTokenizer = t
}

// AnalyzeResult は入力語の形態素解析結果。
type AnalyzeResult struct {
	// Reading は入力語をひらがなに正規化した読み。
	Reading string
	// Known は辞書引き上、既知語(未知語処理にフォールバックしていない)かどうか。
	Known bool
}

// Analyze は入力語(漢字/カタカナ/ひらがな混在可)を形態素解析し、
// ひらがな読みと辞書存在チェック結果を返す(DESIGN.md 4.1・4.2・8章参照)。
//
// IPADICの見出し語はひらがな表記のみ(例: 「しりとり」)、カタカナ表記のみ
// (例: 「パン」のような外来語)のいずれかで登録されていることが多く、表記が
// 揺れると同じ単語でも既知語と判定されないことがある。そのため、入力そのままで
// 未知語と判定された場合は、ひらがな/カタカナを入れ替えた表記でも解析を試みる。
func Analyze(word string) AnalyzeResult {
	result := analyzeSurface(word)
	if result.Known {
		return result
	}

	alt := toggleKanaScript(word)
	if alt == word {
		return result
	}
	altResult := analyzeSurface(alt)
	if altResult.Known {
		return altResult
	}

	return result
}

func analyzeSurface(word string) AnalyzeResult {
	tokens := sharedTokenizer.Analyze(word, tokenizer.Normal)

	var readingKatakana strings.Builder
	known := len(tokens) > 0
	for _, tok := range tokens {
		reading, ok := tok.Reading()
		if !ok || reading == "*" {
			// 未知語トークンは読みを持たないことがある。表層形をそのまま読みとして扱う。
			reading = tok.Surface
		}
		readingKatakana.WriteString(reading)

		if tok.Class != tokenizer.KNOWN {
			known = false
		}
	}

	return AnalyzeResult{
		Reading: katakanaToHiragana(readingKatakana.String()),
		Known:   known,
	}
}

// toggleKanaScript はひらがなをカタカナに、カタカナをひらがなに入れ替える。
// 漢字等それ以外の文字はそのまま残す。
func toggleKanaScript(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 0x3041 && r <= 0x3096: // ひらがな → カタカナ
			b.WriteRune(r + 0x60)
		case r >= 0x30A1 && r <= 0x30F6: // カタカナ → ひらがな
			b.WriteRune(r - 0x60)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
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
