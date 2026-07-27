package auth

import (
	auth "CabinetREST/internal/jwt"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type LoginRequest struct {
	LoginName string `json:"login_name"`
	Password  string `json:"password"`
}

type UserProvider interface {
	GetUserByLogin(
		ctx context.Context,
		loginName string,
	) (User, error)
}

func NewLogin(log *slog.Logger, userProvider UserProvider, tokenService *auth.TokenService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		req.LoginName = strings.TrimSpace(req.LoginName)

		if req.LoginName == "" || req.Password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "login_name and password are required"})
			return
		}

		user, err := userProvider.GetUserByLogin(r.Context(), req.LoginName)
		if err != nil {
			if !errors.Is(err, ErrUserNotFound) {
				log.Error("failed to get user", slog.String("error", err.Error()))
			}

			writeInvalidCredentials(w)
			return
		}

		if !user.Enabled {
			writeInvalidCredentials(w)
			return
		}

		token, expiresAt, err := tokenService.CreateToken(user.ID, user.LoginName)
		if err != nil {
			log.Error("failed to create JWT", slog.String("error", err.Error()))
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_at":   expiresAt,
			"user_id":      user.ID,
			"login_name":   user.LoginName,
		})
	}
}

func writeInvalidCredentials(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, map[string]string{
		"error": "invalid login or password",
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}
