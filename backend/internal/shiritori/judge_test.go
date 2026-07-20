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
		{"東京", "とうきょう"},
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
