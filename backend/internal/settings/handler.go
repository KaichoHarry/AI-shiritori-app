package settings

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/auth"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Routes(requireAuth func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(requireAuth)
	r.Get("/", h.get)
	r.Put("/", h.update)
	return r
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	s, err := h.repo.Get(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "設定の取得に失敗しました")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

type updateRequest struct {
	Mode2Difficulty Difficulty `json:"mode2_difficulty"`
	Mode3Tone       Tone       `json:"mode3_tone"`
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "リクエストの形式が不正です")
		return
	}
	if !isValidDifficulty(req.Mode2Difficulty) || !isValidTone(req.Mode3Tone) {
		writeError(w, http.StatusBadRequest, "invalid_value", "difficulty または tone の値が不正です")
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	s, err := h.repo.Update(r.Context(), userID, req.Mode2Difficulty, req.Mode3Tone)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "設定の更新に失敗しました")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func isValidDifficulty(d Difficulty) bool {
	switch d {
	case DifficultyEasy, DifficultyNormal, DifficultyHard:
		return true
	}
	return false
}

func isValidTone(t Tone) bool {
	switch t {
	case ToneFriendly, TonePolite, ToneComedy:
		return true
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, reason, message string) {
	writeJSON(w, status, map[string]any{"reason": reason, "message": message})
}
