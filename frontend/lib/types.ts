export type User = {
  id: string;
  email: string;
  display_name: string;
};

export type Difficulty = "easy" | "normal" | "hard";
export type Tone = "friendly" | "polite" | "comedy";

export type Settings = {
  user_id: string;
  mode2_difficulty: Difficulty;
  mode3_tone: Tone;
  updated_at: string;
};

export type Mode = "solo" | "vs_ai" | "free_talk";
export type SessionStatus = "in_progress" | "finished";
export type SessionResult =
  | "win"
  | "lose"
  | "ended_by_n"
  | "ended_by_duplicate"
  | "ended_by_user";

export type Session = {
  id: string;
  mode: Mode;
  difficulty?: Difficulty;
  tone?: Tone;
  status: SessionStatus;
  result?: SessionResult;
  turn_count: number;
  started_at: string;
  ended_at?: string | null;
};

export type WordRecord = {
  seq_no: number;
  speaker: "user" | "ai";
  word: string;
  reading: string;
  created_at: string;
};

export type MessageRecord = {
  seq_no: number;
  speaker: "user" | "ai";
  content: string;
  ending_sound: string;
  created_at: string;
};

export type WordView = {
  word: string;
  reading: string;
};

export type NextHint = {
  hiragana: string;
  katakana: string;
};

export type WordSubmitResult = {
  accepted: boolean;
  fatal: boolean;
  reason?: string;
  message?: string;
  player_word?: WordView;
  ai_word?: WordView;
  status: SessionStatus;
  result?: SessionResult;
  next_hint?: NextHint;
};

export type MessageView = {
  content: string;
  ending_sound: string;
};

export type MessageSubmitResult = {
  user_message: MessageView;
  ai_message?: MessageView;
  status: SessionStatus;
  result?: SessionResult;
  next_hint?: NextHint;
};

export type ApiError = {
  reason: string;
  message: string;
};
