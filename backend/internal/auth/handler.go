package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type contextKey string

const userIDContextKey contextKey = "userID"

type Handler struct {
	service *Service
	tokens  *TokenIssuer
}

func NewHandler(service *Service, tokens *TokenIssuer) *Handler {
	return &Handler{service: service, tokens: tokens}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)
	r.Post("/refresh", h.refresh)
	r.Post("/password-reset/request", h.requestPasswordReset)
	r.With(h.RequireAuth).Get("/me", h.me)
	return r
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "認証情報がありません")
			return
		}
		userID, err := h.tokens.ParseAccessToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "認証情報が無効です")
			return
		}
		ctx := withUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "リクエストの形式が不正です")
		return
	}
	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", "email, password, display_name は必須です")
		return
	}

	u, err := h.service.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if errors.Is(err, ErrEmailAlreadyRegistered) {
		writeError(w, http.StatusConflict, "email_already_registered", "このメールアドレスは既に登録されています")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "登録に失敗しました")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":           u.ID,
		"email":        u.Email,
		"display_name": u.DisplayName,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "リクエストの形式が不正です")
		return
	}

	u, pair, err := h.service.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "メールアドレスまたはパスワードが違います")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "ログインに失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"user": map[string]any{
			"id":           u.ID,
			"email":        u.Email,
			"display_name": u.DisplayName,
		},
	})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	// JWTはステートレスなため、サーバー側での失効処理は行わない(暫定案)。
	// クライアント側でトークンを破棄することでログアウトを実現する。
	w.WriteHeader(http.StatusNoContent)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "リクエストの形式が不正です")
		return
	}

	pair, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_refresh_token", "リフレッシュトークンが無効です")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	u, err := h.service.Me(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "ユーザーが見つかりません")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":           u.ID,
		"email":        u.Email,
		"display_name": u.DisplayName,
	})
}

type requestPasswordResetRequest struct {
	Email string `json:"email"`
}

func (h *Handler) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req requestPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "リクエストの形式が不正です")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", "email は必須です")
		return
	}

	if err := h.service.RequestPasswordReset(r.Context(), req.Email); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "申請に失敗しました")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"message": "管理者の承認をお待ちください。承認され次第、登録済みのメールアドレスに新しいパスワードが送信されます。",
	})
}
