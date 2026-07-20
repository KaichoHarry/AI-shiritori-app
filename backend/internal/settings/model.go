package settings

import "time"

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyNormal Difficulty = "normal"
	DifficultyHard   Difficulty = "hard"
)

type Tone string

const (
	ToneFriendly Tone = "friendly"
	TonePolite   Tone = "polite"
	ToneComedy   Tone = "comedy"
)

type Settings struct {
	UserID          string     `json:"user_id"`
	Mode2Difficulty Difficulty `json:"mode2_difficulty"`
	Mode3Tone       Tone       `json:"mode3_tone"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
