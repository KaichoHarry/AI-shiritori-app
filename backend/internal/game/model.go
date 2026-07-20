package game

import "time"

type Mode string

const (
	ModeSolo     Mode = "solo"
	ModeVsAI     Mode = "vs_ai"
	ModeFreeTalk Mode = "free_talk"
)

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusFinished   Status = "finished"
)

// Result はセッションの結果。モード1/2は win/lose、モード3はended_by_*系(DESIGN.md 6.2節)。
type Result string

const (
	ResultWin              Result = "win"
	ResultLose             Result = "lose"
	ResultEndedByN         Result = "ended_by_n"
	ResultEndedByDuplicate Result = "ended_by_duplicate"
	ResultEndedByUser      Result = "ended_by_user"
)

type Session struct {
	ID         string
	UserID     string
	Mode       Mode
	Difficulty *string
	Tone       *string
	Status     Status
	Result     *Result
	TurnCount  int
	StartedAt  time.Time
	EndedAt    *time.Time
}

type Speaker string

const (
	SpeakerUser Speaker = "user"
	SpeakerAI   Speaker = "ai"
)

// Word はモード1・2の使用済み単語(game_words)。
type Word struct {
	ID        string
	SessionID string
	SeqNo     int
	Speaker   Speaker
	Word      string
	Reading   string
	CreatedAt time.Time
}

// MessageRecord はモード3の発言履歴(game_messages)。
type MessageRecord struct {
	ID          string
	SessionID   string
	SeqNo       int
	Speaker     Speaker
	Content     string
	EndingSound string
	CreatedAt   time.Time
}
