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
// 読み(ひらがな)の先頭モーラ別に索引化したもの。Gemini APIを使わずに
// 決定的・無料でAIの手番を選べるようにする(DESIGN.md 5.1節の代替実装)。
var wordBank = buildWordBank()

func buildWordBank() map[string][]string {
	d := ipa.Dict()
	readingIdx, ok := d.ContentsMeta[dict.ReadingIndex]
	if !ok {
		return map[string][]string{}
	}

	buckets := map[string]map[string]struct{}{}
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
		if EndsWithN(hiragana) {
			// 「ん」で終わる語はしりとりで即敗北になるだけなので候補から除外する。
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
	return result
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
