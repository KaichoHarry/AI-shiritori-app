package shiritori

import "testing"

func TestRandomWord(t *testing.T) {
	word, ok := RandomWord("り", map[string]bool{})
	if !ok {
		t.Fatal("expected a word starting with り")
	}
	if FirstSound(word) != "り" {
		t.Errorf("RandomWord(り) = %q, first sound = %q, want り", word, FirstSound(word))
	}
	if EndsWithN(word) {
		t.Errorf("RandomWord returned a word ending in ん: %q", word)
	}

	// 実際にJudgeを通して、生成された単語がそのまま受理されることを確認する
	// (辞書由来なので既知語のはず)。
	result := Judge(word, "り", map[string]bool{})
	if !result.Accepted {
		t.Errorf("Judge(%q) not accepted: %+v", word, result)
	}
}

func TestRandomWordExcludesUsed(t *testing.T) {
	first, ok := RandomWord("り", map[string]bool{})
	if !ok {
		t.Fatal("expected a word starting with り")
	}
	exclude := map[string]bool{first: true}

	for range 50 {
		word, ok := RandomWord("り", exclude)
		if !ok {
			t.Fatal("expected another word starting with り")
		}
		if word == first {
			t.Fatalf("RandomWord returned an excluded word: %q", word)
		}
	}
}

func TestRandomWordNotFoundForRareSound(t *testing.T) {
	// 「ん」で始まる語は日本語にほぼ存在しないため、除外プールが尽きるケースを検証する。
	all := wordBank["ん"]
	exclude := map[string]bool{}
	for _, w := range all {
		exclude[w] = true
	}
	if _, ok := RandomWord("ん", exclude); ok {
		t.Fatal("expected RandomWord to report exhaustion when all candidates are excluded")
	}
}
