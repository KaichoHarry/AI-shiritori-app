package shiritori

import "testing"

func TestAnalyzeKnownWords(t *testing.T) {
	cases := []struct {
		word    string
		reading string
	}{
		{"しりとり", "しりとり"},
		{"りんご", "りんご"},
		{"コーヒー", "こーひー"},
		// 「花火」はKnown=trueだが「はなび」(ひらがな)は未知語判定になるという
		// 不整合があった。読みの集合を直接引く方式に変えたことで正しく実在語と認識される。
		{"はなび", "はなび"},
		{"ハナビ", "はなび"},
	}
	for _, c := range cases {
		got := Analyze(c.word)
		if !got.Known {
			t.Errorf("Analyze(%q).Known = false, want true", c.word)
		}
		if got.Reading != c.reading {
			t.Errorf("Analyze(%q).Reading = %q, want %q", c.word, got.Reading, c.reading)
		}
	}
}

func TestAnalyzeUnknownWord(t *testing.T) {
	got := Analyze("ぎゃばんずぽんてけ")
	if got.Known {
		t.Errorf("Analyze(nonsense word).Known = true, want false (got reading %q)", got.Reading)
	}
}

// TestAnalyzeSingleSmallKana は、小さい仮名1文字だけの入力が全て未知語として
// 拒否されることを検証する(「ぁ」「っ」がIPADIC辞書内の記号的エントリを介して
// 誤って既知語判定されていた問題の再発防止)。
func TestAnalyzeSingleSmallKana(t *testing.T) {
	smallKanaChars := []string{
		"ぁ", "ぃ", "ぅ", "ぇ", "ぉ", "ゃ", "ゅ", "ょ", "っ", "ゎ",
		"ァ", "ィ", "ゥ", "ェ", "ォ", "ャ", "ュ", "ョ", "ッ", "ヮ",
	}
	for _, c := range smallKanaChars {
		got := Analyze(c)
		if got.Known {
			t.Errorf("Analyze(%q).Known = true, want false (single small kana should never be a valid word)", c)
		}
	}
}

func TestLastSoundLongVowel(t *testing.T) {
	// DESIGN.md 1章の例: 「コーヒー」→「ひ」
	got := LastSound("こーひー")
	if got != "ひ" {
		t.Errorf("LastSound(こーひー) = %q, want %q", got, "ひ")
	}
}

func TestFirstSound(t *testing.T) {
	if got := FirstSound("しゃしん"); got != "しゃ" {
		t.Errorf("FirstSound(しゃしん) = %q, want %q", got, "しゃ")
	}
	if got := FirstSound("りんご"); got != "り" {
		t.Errorf("FirstSound(りんご) = %q, want %q", got, "り")
	}
	// 促音「っ」は拗音と異なり、直前の文字と結合せず独立した1モーラとして扱う。
	if got := FirstSound("らっぱ"); got != "ら" {
		t.Errorf("FirstSound(らっぱ) = %q, want %q", got, "ら")
	}
}

func TestLastSoundForeignSmallVowel(t *testing.T) {
	// 外来語表記の小さい母音(ぁぃぅぇぉ)は拗音と同様に直前の文字と結合するため、
	// 「カフェ」の末尾モーラは独立した「ぇ」ではなく「ふぇ」になる。
	cases := []struct{ reading, want string }{
		{"かふぇ", "ふぇ"},
		{"ぱーてぃー", "てぃ"},
		{"れでぃー", "でぃ"},
		{"そふぁー", "ふぁ"},
	}
	for _, c := range cases {
		if got := LastSound(c.reading); got != c.want {
			t.Errorf("LastSound(%q) = %q, want %q", c.reading, got, c.want)
		}
	}
}

func TestEndsWithN(t *testing.T) {
	if !EndsWithN("ぱん") {
		t.Errorf("EndsWithN(ぱん) = false, want true")
	}
	if EndsWithN("ぱんだ") {
		t.Errorf("EndsWithN(ぱんだ) = true, want false")
	}
}

func TestJudgeAcceptsFirstWord(t *testing.T) {
	result := Judge("しりとり", "", map[string]bool{})
	if !result.Accepted {
		t.Fatalf("Judge(しりとり) not accepted: %+v", result)
	}
	if result.Reading != "しりとり" {
		t.Errorf("Reading = %q, want しりとり", result.Reading)
	}
}

func TestJudgeInvalidCharacters(t *testing.T) {
	cases := []string{"理科", "apple", "しりとり1", "パン。"}
	for _, word := range cases {
		result := Judge(word, "", map[string]bool{})
		if result.Accepted {
			t.Errorf("Judge(%q) accepted, want rejection", word)
		}
		if result.Fatal {
			t.Errorf("Judge(%q) fatal, want non-fatal", word)
		}
		if result.Reason != ReasonInvalidCharacters {
			t.Errorf("Judge(%q).Reason = %q, want %q", word, result.Reason, ReasonInvalidCharacters)
		}
	}
}

func TestJudgeWordNotFound(t *testing.T) {
	result := Judge("ぎゃばんずぽんてけ", "", map[string]bool{})
	if result.Accepted {
		t.Fatalf("expected not accepted for nonsense word")
	}
	if result.Fatal {
		t.Errorf("word_not_found should not be fatal")
	}
	if result.Reason != ReasonWordNotFound {
		t.Errorf("Reason = %q, want %q", result.Reason, ReasonWordNotFound)
	}
}

func TestJudgeEndsWithN(t *testing.T) {
	result := Judge("ぱん", "", map[string]bool{})
	if result.Accepted || !result.Fatal || result.Reason != ReasonEndsWithN {
		t.Fatalf("unexpected result for ぱん: %+v", result)
	}
}

func TestJudgeConnectionMismatch(t *testing.T) {
	// 直前の単語「りんご」の末尾音は「ご」なので、「たぬき」(た始まり)は繋がらない。
	result := Judge("たぬき", "りんご", map[string]bool{})
	if result.Accepted || !result.Fatal || result.Reason != ReasonConnectionMismatch {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestJudgeConnectionMatch(t *testing.T) {
	result := Judge("ごりら", "りんご", map[string]bool{})
	if !result.Accepted {
		t.Fatalf("expected acceptance for ごりら following りんご: %+v", result)
	}
}

func TestJudgeConnectionMatchAcrossSokuon(t *testing.T) {
	// 「ごりら」(末尾「ら」)→「らっぱ」は一般的なしりとりで成立する(ユーザー確認済み)。
	result := Judge("らっぱ", "ごりら", map[string]bool{})
	if !result.Accepted {
		t.Fatalf("expected acceptance for らっぱ following ごりら: %+v", result)
	}
}

func TestJudgeConnectionMatchWithLongVowel(t *testing.T) {
	// 直前が「コーヒー」(末尾音「ひ」)の場合、「ひこうき」は「ひ」始まりなので繋がる。
	result := Judge("ひこうき", "こーひー", map[string]bool{})
	if !result.Accepted {
		t.Fatalf("expected acceptance for ひこうき following こーひー: %+v", result)
	}
}

func TestJudgeDuplicateWord(t *testing.T) {
	used := map[string]bool{"りんご": true}
	result := Judge("りんご", "", used)
	if result.Accepted || !result.Fatal || result.Reason != ReasonDuplicateWord {
		t.Fatalf("unexpected result: %+v", result)
	}
}
