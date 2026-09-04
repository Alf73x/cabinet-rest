package auth

import (
	auth "CabinetREST/internal/jwt"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	maxLoginBodyBytes   = 16 << 10
	maxLoginNameBytes   = 128
	maxPasswordBytes    = 256
	loginRateWindow     = time.Minute
	maxAttemptsPerIP    = 10
	maxAttemptsPerUser  = 5
	maxRateLimitEntries = 4096
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
	limiter := newLoginRateLimiter(loginRateWindow)
	initDummyPasswordHash()

	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest

		r.Body = http.MaxBytesReader(w, r.Body, maxLoginBodyBytes)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		err := decoder.Decode(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		req.LoginName = strings.TrimSpace(req.LoginName)

		if req.LoginName == "" || req.Password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "login_name and password are required"})
			return
		}
		if len(req.LoginName) > maxLoginNameBytes || len(req.Password) > maxPasswordBytes {
			writeInvalidCredentials(w)
			return
		}

		if !limiter.Allow(clientIP(r), strings.ToLower(req.LoginName)) {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many login attempts"})
			return
		}

		user, err := userProvider.GetUserByLogin(r.Context(), req.LoginName)
		if err != nil {
			CheckDummyPassword(req.Password)
			if !errors.Is(err, ErrUserNotFound) {
				log.Error("failed to get user", slog.String("error", err.Error()))
			}

			writeInvalidCredentials(w)
			return
		}

		// Проверяем хеш даже для отключённого пользователя, чтобы время ответа
		// не выдавало состояние учётной записи.
		if !CheckPassword(user.PasswordHash, req.Password) || !user.Enabled {
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

type loginAttempt struct {
	startedAt time.Time
	count     int
}

type loginRateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	attempts map[string]loginAttempt
}

func newLoginRateLimiter(window time.Duration) *loginRateLimiter {
	return &loginRateLimiter{window: window, attempts: make(map[string]loginAttempt)}
}

func (l *loginRateLimiter) Allow(ip, loginName string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if len(l.attempts) >= maxRateLimitEntries {
		for key, attempt := range l.attempts {
			if now.Sub(attempt.startedAt) >= l.window {
				delete(l.attempts, key)
			}
		}
	}
	keys := []struct {
		key   string
		limit int
	}{
		{key: "ip:" + ip, limit: maxAttemptsPerIP},
		{key: "login:" + loginName, limit: maxAttemptsPerUser},
	}
	for _, item := range keys {
		if _, exists := l.attempts[item.key]; !exists && len(l.attempts) >= maxRateLimitEntries {
			return false
		}
	}

	allowed := true
	for _, item := range keys {
		attempt := l.attempts[item.key]
		if attempt.startedAt.IsZero() || now.Sub(attempt.startedAt) >= l.window {
			attempt = loginAttempt{startedAt: now}
		}
		if attempt.count >= item.limit {
			allowed = false
		}
	}
	if allowed {
		for _, item := range keys {
			attempt := l.attempts[item.key]
			if attempt.startedAt.IsZero() || now.Sub(attempt.startedAt) >= l.window {
				attempt = loginAttempt{startedAt: now}
			}
			attempt.count++
			l.attempts[item.key] = attempt
		}
	}

	return allowed
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
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
