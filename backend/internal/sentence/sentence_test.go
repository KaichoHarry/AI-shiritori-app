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
		{"ねこも気持ちよさそうに寝てるよ", "よ"},
		{"布団が吹っ飛んだ", "だ"},
	}
	for _, c := range cases {
		got := ExtractEndingSound(c.content)
		if got != c.want {
			t.Errorf("ExtractEndingSound(%q) = %q, want %q", c.content, got, c.want)
		}
	}
}

func TestEvaluateEndsWithN(t *testing.T) {
	result := Evaluate("それはとても残念", nil)
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
