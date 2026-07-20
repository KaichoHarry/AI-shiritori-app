package game

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, reason, message string) {
	writeJSON(w, status, map[string]any{"reason": reason, "message": message})
}

func sessionView(s *Session) map[string]any {
	v := map[string]any{
		"id":         s.ID,
		"mode":       s.Mode,
		"status":     s.Status,
		"turn_count": s.TurnCount,
		"started_at": s.StartedAt,
		"ended_at":   s.EndedAt,
	}
	if s.Difficulty != nil {
		v["difficulty"] = *s.Difficulty
	}
	if s.Tone != nil {
		v["tone"] = *s.Tone
	}
	if s.Result != nil {
		v["result"] = *s.Result
	}
	return v
}

func wordRecordViews(words []*Word) []map[string]any {
	views := make([]map[string]any, 0, len(words))
	for _, w := range words {
		views = append(views, map[string]any{
			"seq_no":     w.SeqNo,
			"speaker":    w.Speaker,
			"word":       w.Word,
			"reading":    w.Reading,
			"created_at": w.CreatedAt,
		})
	}
	return views
}

func messageRecordViews(messages []*MessageRecord) []map[string]any {
	views := make([]map[string]any, 0, len(messages))
	for _, m := range messages {
		views = append(views, map[string]any{
			"seq_no":       m.SeqNo,
			"speaker":      m.Speaker,
			"content":      m.Content,
			"ending_sound": m.EndingSound,
			"created_at":   m.CreatedAt,
		})
	}
	return views
}

func wordView(v *WordView) map[string]any {
	if v == nil {
		return nil
	}
	return map[string]any{"word": v.Word, "reading": v.Reading}
}

func wordResultView(r *WordResult) map[string]any {
	v := map[string]any{
		"accepted": r.Accepted,
		"fatal":    r.Fatal,
		"status":   r.Status,
	}
	if r.Reason != ReasonNone {
		v["reason"] = r.Reason
	}
	if r.Message != "" {
		v["message"] = r.Message
	}
	if r.PlayerWord != nil {
		v["player_word"] = wordView(r.PlayerWord)
	}
	if r.AIWord != nil {
		v["ai_word"] = wordView(r.AIWord)
	}
	if r.Result != nil {
		v["result"] = *r.Result
	}
	return v
}

func messageViewJSON(v *MessageView) map[string]any {
	if v == nil {
		return nil
	}
	return map[string]any{"content": v.Content, "ending_sound": v.EndingSound}
}

func messageResultView(r *MessageResult) map[string]any {
	v := map[string]any{
		"status":       r.Status,
		"user_message": messageViewJSON(r.UserMessage),
	}
	if r.AIMessage != nil {
		v["ai_message"] = messageViewJSON(r.AIMessage)
	}
	if r.Result != nil {
		v["result"] = *r.Result
	}
	return v
}
