package game

import "testing"

func TestShouldAIIntentionallyFailGracePeriod(t *testing.T) {
	// normal難易度は49ラリー目まで(境界値含む)、確率に関わらず絶対に失敗しない。
	for rally := 1; rally <= aiNormalDifficultyGracePeriodRally; rally++ {
		if shouldAIIntentionallyFail("normal", rally) {
			t.Fatalf("normal rally=%d: expected no intentional failure within grace period", rally)
		}
	}
}

func TestShouldAIIntentionallyFailHardNeverFails(t *testing.T) {
	// hard難易度は確率的な失敗が一切無い(辞書が尽きた場合のみ別経路で負ける)。
	for rally := 1; rally <= 200; rally++ {
		if shouldAIIntentionallyFail("hard", rally) {
			t.Fatalf("hard rally=%d: expected hard difficulty to never intentionally fail", rally)
		}
	}
}

func TestShouldAIIntentionallyFailEasyRoughlyOnePercent(t *testing.T) {
	const trials = 200000
	failures := 0
	for i := 0; i < trials; i++ {
		if shouldAIIntentionallyFail("easy", 1) {
			failures++
		}
	}
	rate := float64(failures) / float64(trials)
	if rate < 0.005 || rate > 0.015 {
		t.Errorf("easy failure rate = %.4f, want close to 0.01", rate)
	}
}

func TestShouldAIIntentionallyFailNormalAfterGracePeriod(t *testing.T) {
	const trials = 200000
	failures := 0
	for i := 0; i < trials; i++ {
		if shouldAIIntentionallyFail("normal", aiNormalDifficultyGracePeriodRally+1) {
			failures++
		}
	}
	rate := float64(failures) / float64(trials)
	if rate < 0.005 || rate > 0.015 {
		t.Errorf("normal (post grace period) failure rate = %.4f, want close to 0.01", rate)
	}
}
