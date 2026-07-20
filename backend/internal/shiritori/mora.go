package shiritori

// smallKana は直前の文字と結合して1音として扱う小さい文字(DESIGN.md 1章参照)。
var smallKana = map[rune]bool{
	'っ': true,
	'ゃ': true,
	'ゅ': true,
	'ょ': true,
}

const longVowelMark = 'ー'

// splitIntoMorae はひらがな読みを、しりとり判定上の「音」の単位(モーラ)に分割する。
//
//   - 拗音(っ・ゃ・ゅ・ょ)は直前の文字と結合して1モーラとして扱う。
//   - 長音「ー」は直前のモーラに吸収され、比較対象の文字列には含めない
//     (例: 「コーヒー」→ [こ, ひ]。末尾の「ー」は直前の母音音を使うというDESIGN.md 1章の
//     ルールに従い、末尾モーラは「ひ」になる)。
func splitIntoMorae(reading string) []string {
	var morae []string
	for _, r := range reading {
		switch {
		case r == longVowelMark && len(morae) > 0:
			// 長音は直前のモーラに吸収し、文字列としては付け足さない。
		case smallKana[r] && len(morae) > 0:
			morae[len(morae)-1] += string(r)
		default:
			morae = append(morae, string(r))
		}
	}
	return morae
}

// FirstSound は読みの最初のモーラを返す。
func FirstSound(reading string) string {
	morae := splitIntoMorae(reading)
	if len(morae) == 0 {
		return ""
	}
	return morae[0]
}

// LastSound は読みの最後のモーラを返す(長音の扱いはDESIGN.md 1章参照)。
func LastSound(reading string) string {
	morae := splitIntoMorae(reading)
	if len(morae) == 0 {
		return ""
	}
	return morae[len(morae)-1]
}

// EndsWithN は読みが「ん」で終わっているかどうかを判定する。
func EndsWithN(reading string) bool {
	rs := []rune(reading)
	if len(rs) == 0 {
		return false
	}
	return rs[len(rs)-1] == 'ん'
}
