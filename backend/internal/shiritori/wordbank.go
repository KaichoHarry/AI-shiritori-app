package shiritori

import (
	"math/rand"

	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/ipa"
)

// excludedNounSubcategories はしりとりの単語として不適切な名詞細分類
// (接尾: 「的」「さん」等、単独の語として成立しない)。
var excludedNounSubcategories = map[string]bool{
	"接尾": true,
}

// wordBank はモード2(AI対戦)のAI側の単語選択用に、IPADIC辞書の名詞を
// 読み(ひらがな)の先頭モーラ別に索引化したもの(「ん」で終わる語は除外)。
// Gemini APIを使わずに決定的・無料でAIの手番を選べるようにする(DESIGN.md 5.1節の代替実装)。
//
// knownReadings はプレイヤー入力の辞書存在チェック用に、同じ名詞集合を読み(ひらがな)の
// 集合として保持したもの(「ん」で終わる語も含む。「ん」終端の判定はJudge内の
// 別ステップ(EndsWithN)が担当するため、辞書存在チェックの時点では除外しない)。
//
// 以前はプレイヤー入力の辞書チェックにkagomeのトークナイザで再解析する方式
// (Analyze/Known)を使っていたが、「花火」はKnown=trueなのに読みの「はなび」
// (ひらがな)は未知語判定になる、逆にAIが単語バンクから生成した語をプレイヤー入力
// として検証すると辞書に無いと判定される、といった不整合があった。加えて、
// ひらがな⇔カタカナ変換フォールバックの副作用で「ぁ」「っ」のような小さい仮名
// 1文字が(IPADIC辞書の記号的エントリを拾ってしまい)誤って既知語判定されていた。
// 読み(ひらがな)の集合を直接引く方式に統一することで、AIの単語ソースとプレイヤーの
// 辞書チェックの基準を一致させ、これらの問題を同時に解消する。
var wordBank, knownReadings = buildWordBank()

func buildWordBank() (map[string][]string, map[string]bool) {
	d := ipa.Dict()
	readingIdx, ok := d.ContentsMeta[dict.ReadingIndex]
	if !ok {
		return map[string][]string{}, map[string]bool{}
	}

	buckets := map[string]map[string]struct{}{}
	known := map[string]bool{}
	for id := range d.Morphs {
		pos := d.POSTable.POSs[id]
		if len(pos) == 0 || d.POSTable.NameList[pos[0]] != "名詞" {
			continue
		}
		if len(pos) > 1 && excludedNounSubcategories[d.POSTable.NameList[pos[1]]] {
			continue
		}

		reading := knownFeatureAt(d, id, int(readingIdx))
		if reading == "" || reading == "*" {
			continue
		}

		hiragana := katakanaToHiragana(reading)
		known[hiragana] = true

		if EndsWithN(hiragana) {
			// 「ん」で終わる語はAIの候補としては使えないので単語バンクからは除外するが、
			// 辞書存在チェック(knownReadings)には残す。
			continue
		}

		first := FirstSound(hiragana)
		if first == "" {
			continue
		}
		if buckets[first] == nil {
			buckets[first] = map[string]struct{}{}
		}
		buckets[first][hiragana] = struct{}{}
	}

	result := make(map[string][]string, len(buckets))
	for first, set := range buckets {
		list := make([]string, 0, len(set))
		for reading := range set {
			list = append(list, reading)
		}
		result[first] = list
	}
	return result, known
}

// knownFeatureAt はkagomeのToken.FeatureAt相当のロジックを、既知語IDから
// トークン化を経由せず直接引くための実装(POS階層 + Contentsの結合順序に従う)。
func knownFeatureAt(d *dict.Dict, id, i int) string {
	pos := d.POSTable.POSs[id]
	if i < len(pos) {
		posID := pos[i]
		if int(posID) < len(d.POSTable.NameList) {
			return d.POSTable.NameList[posID]
		}
		return ""
	}
	ci := i - len(pos)
	if id >= len(d.Contents) {
		return ""
	}
	c := d.Contents[id]
	if ci < 0 || ci >= len(c) {
		return ""
	}
	return c[ci]
}

// RandomWord は辞書からrequiredFirstSoundで始まり、excludeReadingsに含まれない
// (かつ「ん」で終わらない)単語をランダムに1つ選ぶ。見つからなければok=falseを返す。
func RandomWord(requiredFirstSound string, excludeReadings map[string]bool) (reading string, ok bool) {
	candidates := wordBank[requiredFirstSound]
	if len(candidates) == 0 {
		return "", false
	}

	available := make([]string, 0, len(candidates))
	for _, w := range candidates {
		if !excludeReadings[w] {
			available = append(available, w)
		}
	}
	if len(available) == 0 {
		return "", false
	}
	return available[rand.Intn(len(available))], true
}

// IsKnownReading は、ひらがな正規化済みの読みがIPADIC辞書の名詞として実在するかどうかを
// 判定する(プレイヤー入力の辞書存在チェック用)。
func IsKnownReading(hiraganaReading string) bool {
	return knownReadings[hiraganaReading]
}
