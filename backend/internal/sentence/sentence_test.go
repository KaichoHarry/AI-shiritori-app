package sentence

import "testing"

// DESIGN.md 2.3節の例「今日は天気がいいね」→(文末の読みは「い」)は、実際の文末文字が
// 「ね」であることと矛盾しており、ドキュメント側のtypoと判断した(ユーザー確認済み)。
// 本実装は文字通り最後のモーラを文末音とする。
func TestExtractEndingSound(t *testing.T) {
	cases := []struct {
		content string
		want    string
	}{
		{"今日は天気がいいね", "ね"},
		{"今日は天気がいいね。", "ね"},
		{"今日は天気がいいね!!", "ね"},
		{"今日は天気がいいね！！", "ね"}, // 全角の感嘆符も除外対象
		{"ねこも気持ちよさそうに寝てるよ", "よ"},
		{"布団が吹っ飛んだ", "だ"},
		// 末尾から遡って最初に見つかった有効文字(ひらがな/カタカナ/英字)を使う仕様のため、
		// 末尾が漢字の場合はその手前の仮名までスキップする(ユーザー確認済み)。
		{"これは私の本", "の"},
		{"hello", "o"},  // 半角英字
		{"了解ですＯＫ", "Ｋ"}, // 全角英字
	}
	for _, c := range cases {
		got := ExtractEndingSound(c.content)
		if got != c.want {
			t.Errorf("ExtractEndingSound(%q) = %q, want %q", c.content, got, c.want)
		}
	}
}

func TestEvaluateEndsWithN(t *testing.T) {
	// 「残念」のような漢字表記は末尾音の判定対象外(ひらがな/カタカナ/英字のみ有効)なので、
	// ひらがなで直接「ん」に終わる文で検証する(ユーザー確認済み仕様)。
	result := Evaluate("それはとてもざんねん", nil)
	if !result.Ended || result.Reason != EndReasonN {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestEvaluateDuplicate(t *testing.T) {
	previous := []string{"今日は天気がいいね", "布団が吹っ飛んだ"}
	result := Evaluate("布団が吹っ飛んだ", previous)
	if !result.Ended || result.Reason != EndReasonDuplicate {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestEvaluateContinues(t *testing.T) {
	result := Evaluate("ねこも気持ちよさそうに寝てるよ", []string{"今日は天気がいいね"})
	if result.Ended {
		t.Fatalf("expected continuation, got ended: %+v", result)
	}
	if result.EndingSound != "よ" {
		t.Errorf("EndingSound = %q, want よ", result.EndingSound)
	}
}
