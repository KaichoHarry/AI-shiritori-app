package game

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateSession(ctx context.Context, userID string, mode Mode, difficulty, tone *string) (*Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx, `
		INSERT INTO game_sessions (user_id, mode, difficulty, tone)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, user_id::text, mode, difficulty, tone, status, result, turn_count, started_at, ended_at
	`, userID, mode, difficulty, tone).Scan(
		&s.ID, &s.UserID, &s.Mode, &s.Difficulty, &s.Tone, &s.Status, &s.Result, &s.TurnCount, &s.StartedAt, &s.EndedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) GetSession(ctx context.Context, sessionID, userID string) (*Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, mode, difficulty, tone, status, result, turn_count, started_at, ended_at
		FROM game_sessions WHERE id = $1 AND user_id = $2
	`, sessionID, userID).Scan(
		&s.ID, &s.UserID, &s.Mode, &s.Difficulty, &s.Tone, &s.Status, &s.Result, &s.TurnCount, &s.StartedAt, &s.EndedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) ListSessions(ctx context.Context, userID string, limit, offset int) ([]*Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, user_id::text, mode, difficulty, tone, status, result, turn_count, started_at, ended_at
		FROM game_sessions WHERE user_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.UserID, &s.Mode, &s.Difficulty, &s.Tone, &s.Status, &s.Result, &s.TurnCount, &s.StartedAt, &s.EndedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, &s)
	}
	return sessions, rows.Err()
}

// FinishSession はセッションを終了状態にする。
func (r *Repository) FinishSession(ctx context.Context, sessionID string, result Result) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE game_sessions SET status = 'finished', result = $2, ended_at = now()
		WHERE id = $1
	`, sessionID, result)
	return err
}

// AddWord は使用済み単語を1件追加し、セッションのturn_countを加算する。
func (r *Repository) AddWord(ctx context.Context, sessionID string, seqNo int, speaker Speaker, word, reading string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO game_words (session_id, seq_no, speaker, word, reading)
		VALUES ($1, $2, $3, $4, $5)
	`, sessionID, seqNo, speaker, word, reading)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `UPDATE game_sessions SET turn_count = $2 WHERE id = $1`, sessionID, seqNo)
	return err
}

func (r *Repository) ListWords(ctx context.Context, sessionID string) ([]*Word, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, session_id::text, seq_no, speaker, word, reading, created_at
		FROM game_words WHERE session_id = $1 ORDER BY seq_no ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []*Word
	for rows.Next() {
		var w Word
		if err := rows.Scan(&w.ID, &w.SessionID, &w.SeqNo, &w.Speaker, &w.Word, &w.Reading, &w.CreatedAt); err != nil {
			return nil, err
		}
		words = append(words, &w)
	}
	return words, rows.Err()
}

// CountWordsBySpeaker はそのセッション内で指定した話者が出した単語数を返す
// (モード2の難易度アルゴリズムにおけるAIのラリー数の算出に使う)。
func (r *Repository) CountWordsBySpeaker(ctx context.Context, sessionID string, speaker Speaker) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT count(*) FROM game_words WHERE session_id = $1 AND speaker = $2
	`, sessionID, speaker).Scan(&count)
	return count, err
}

// UsedReadings はそのセッション内で既に使われた単語のreadingの集合を返す(重複チェック用)。
func (r *Repository) UsedReadings(ctx context.Context, sessionID string) (map[string]bool, error) {
	rows, err := r.pool.Query(ctx, `SELECT reading FROM game_words WHERE session_id = $1`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	used := map[string]bool{}
	for rows.Next() {
		var reading string
		if err := rows.Scan(&reading); err != nil {
			return nil, err
		}
		used[reading] = true
	}
	return used, rows.Err()
}

// LastWordReading はセッション内で直近に追加された単語のreadingを返す(最初のターンなら空文字列)。
func (r *Repository) LastWordReading(ctx context.Context, sessionID string) (string, error) {
	var reading string
	err := r.pool.QueryRow(ctx, `
		SELECT reading FROM game_words WHERE session_id = $1 ORDER BY seq_no DESC LIMIT 1
	`, sessionID).Scan(&reading)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return reading, nil
}

// AddMessage は発言を1件追加し、セッションのturn_countを加算する。
func (r *Repository) AddMessage(ctx context.Context, sessionID string, seqNo int, speaker Speaker, content, endingSound string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO game_messages (session_id, seq_no, speaker, content, ending_sound)
		VALUES ($1, $2, $3, $4, $5)
	`, sessionID, seqNo, speaker, content, endingSound)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `UPDATE game_sessions SET turn_count = $2 WHERE id = $1`, sessionID, seqNo)
	return err
}

func (r *Repository) ListMessages(ctx context.Context, sessionID string) ([]*MessageRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, session_id::text, seq_no, speaker, content, ending_sound, created_at
		FROM game_messages WHERE session_id = $1 ORDER BY seq_no ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*MessageRecord
	for rows.Next() {
		var m MessageRecord
		if err := rows.Scan(&m.ID, &m.SessionID, &m.SeqNo, &m.Speaker, &m.Content, &m.EndingSound, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &m)
	}
	return messages, rows.Err()
}

// RecentMessages は直近N件の発言を古い順で返す(Gemini APIへの会話履歴用)。
func (r *Repository) RecentMessages(ctx context.Context, sessionID string, limit int) ([]*MessageRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, session_id::text, seq_no, speaker, content, ending_sound, created_at
		FROM game_messages WHERE session_id = $1 ORDER BY seq_no DESC LIMIT $2
	`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*MessageRecord
	for rows.Next() {
		var m MessageRecord
		if err := rows.Scan(&m.ID, &m.SessionID, &m.SeqNo, &m.Speaker, &m.Content, &m.EndingSound, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// AllContents はそのセッション内の全発言内容を返す(完全一致の重複チェック用)。
func (r *Repository) AllContents(ctx context.Context, sessionID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT content FROM game_messages WHERE session_id = $1`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		contents = append(contents, c)
	}
	return contents, rows.Err()
}
