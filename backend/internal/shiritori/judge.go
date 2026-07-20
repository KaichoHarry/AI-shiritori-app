package shiritori

// Reason は判定結果の理由。DESIGN.md 7.4節のAPIレスポンス仕様と対応する。
type Reason string

const (
	ReasonNone               Reason = ""
	ReasonInvalidCharacters  Reason = "invalid_characters"
	ReasonWordNotFound       Reason = "word_not_found"
	ReasonEndsWithN          Reason = "ends_with_n"
	ReasonConnectionMismatch Reason = "connection_mismatch"
	ReasonDuplicateWord      Reason = "duplicate_word"
)

// Result はJudgeの判定結果。
type Result struct {
	// Reading は入力語のひらがな正規化読み。
	Reading string
	// Accepted はこのターンの入力として受理され、継続扱いになるかどうか。
	Accepted bool
	// Fatal はゲームオーバー(またはセッション終了)につながる致命的エラーかどうか。
	// Accepted=false かつ Fatal=false の場合は、辞書に無い語のように
	// 同じターンでの再入力を促すべき非致命的エラーを意味する。
	Fatal  bool
	Reason Reason
}

// Judge はDESIGN.md 8章の判定ロジックを実装する、モード1・2共通のしりとり判定関数。
//
// previousReading は直前の単語のひらがな読み(セッション最初の入力の場合は空文字列を渡し、
// 接続チェックをスキップする)。usedReadings はそのセッション内で既に使用済みの読みの集合。
func Judge(word string, previousReading string, usedReadings map[string]bool) Result {
	if !IsKanaOnly(word) {
		return Result{Reading: "", Accepted: false, Fatal: false, Reason: ReasonInvalidCharacters}
	}

	analyzed := Analyze(word)

	if !analyzed.Known {
		return Result{Reading: analyzed.Reading, Accepted: false, Fatal: false, Reason: ReasonWordNotFound}
	}

	if EndsWithN(analyzed.Reading) {
		return Result{Reading: analyzed.Reading, Accepted: false, Fatal: true, Reason: ReasonEndsWithN}
	}

	if previousReading != "" && FirstSound(analyzed.Reading) != LastSound(previousReading) {
		return Result{Reading: analyzed.Reading, Accepted: false, Fatal: true, Reason: ReasonConnectionMismatch}
	}

	if usedReadings[analyzed.Reading] {
		return Result{Reading: analyzed.Reading, Accepted: false, Fatal: true, Reason: ReasonDuplicateWord}
	}

	return Result{Reading: analyzed.Reading, Accepted: true, Fatal: false, Reason: ReasonNone}
}
