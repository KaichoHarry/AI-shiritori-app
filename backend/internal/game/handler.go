package game

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(requireAuth func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(requireAuth)
	r.Post("/", h.createSession)
	r.Get("/", h.listSessions)
	r.Get("/{id}", h.getSession)
	r.Post("/{id}/end", h.endSession)
	r.Post("/{id}/words", h.submitWord)
	r.Post("/{id}/messages", h.submitMessage)
	return r
}

type createSessionRequest struct {
	Mode       string `json:"mode"`
	Difficulty string `json:"difficulty"`
	Tone       string `json:"tone"`
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "リクエストの形式が不正です")
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	var difficulty, tone *string
	if req.Difficulty != "" {
		difficulty = &req.Difficulty
	}
	if req.Tone != "" {
		tone = &req.Tone
	}

	session, err := h.service.CreateSession(r.Context(), userID, Mode(req.Mode), difficulty, tone)
	if errors.Is(err, ErrInvalidMode) {
		writeError(w, http.StatusBadRequest, "invalid_mode", "mode は solo, vs_ai, free_talk のいずれかを指定してください")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "セッションの作成に失敗しました")
		return
	}

	writeJSON(w, http.StatusCreated, sessionView(session))
}

func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	userID := auth.UserIDFromContext(r.Context())
	sessions, err := h.service.ListSessions(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "セッション一覧の取得に失敗しました")
		return
	}

	views := make([]map[string]any, 0, len(sessions))
	for _, s := range sessions {
		views = append(views, sessionView(s))
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": views})
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	session, err := h.service.GetSession(r.Context(), sessionID, userID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "セッションが見つかりません")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "セッションの取得に失敗しました")
		return
	}

	body := map[string]any{"session": sessionView(session)}

	if session.Mode == ModeFreeTalk {
		messages, err := h.service.GetMessages(r.Context(), sessionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "会話履歴の取得に失敗しました")
			return
		}
		body["messages"] = messageRecordViews(messages)
	} else {
		words, err := h.service.GetWords(r.Context(), sessionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "単語履歴の取得に失敗しました")
			return
		}
		body["words"] = wordRecordViews(words)
	}

	writeJSON(w, http.StatusOK, body)
}

func (h *Handler) endSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	session, err := h.service.EndSession(r.Context(), sessionID, userID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "セッションが見つかりません")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "セッションの終了に失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, sessionView(session))
}

type submitWordRequest struct {
	Word string `json:"word"`
}

func (h *Handler) submitWord(w http.ResponseWriter, r *http.Request) {
	var req submitWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
		writeError(w, http.StatusBadRequest, "invalid_body", "word は必須です")
		return
	}

	sessionID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	result, err := h.service.SubmitWord(r.Context(), sessionID, userID, req.Word)
	if handledGameError(w, err) {
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "判定処理に失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, wordResultView(result))
}

type submitMessageRequest struct {
	Content string `json:"content"`
}

func (h *Handler) submitMessage(w http.ResponseWriter, r *http.Request) {
	var req submitMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		writeError(w, http.StatusBadRequest, "invalid_body", "content は必須です")
		return
	}

	sessionID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	result, err := h.service.SubmitMessage(r.Context(), sessionID, userID, req.Content)
	if handledGameError(w, err) {
		return
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, "ai_generation_failed", "AIの応答生成に失敗しました。もう一度お試しください")
		return
	}

	writeJSON(w, http.StatusOK, messageResultView(result))
}

// handledGameError はハンドラー共通のエラーをレスポンスに変換する。処理済みならtrueを返す。
func handledGameError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "セッションが見つかりません")
		return true
	case errors.Is(err, ErrWrongModeForCall):
		writeError(w, http.StatusBadRequest, "wrong_mode", "このセッションのモードではこの操作はできません")
		return true
	case errors.Is(err, ErrSessionFinished):
		writeError(w, http.StatusConflict, "session_finished", "このセッションは既に終了しています")
		return true
	default:
		return false
	}
}
