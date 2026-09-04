package auth

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	jwtservice "CabinetREST/internal/jwt"
)

type loginUserProviderFunc func(context.Context, string) (User, error)

func (f loginUserProviderFunc) GetUserByLogin(ctx context.Context, loginName string) (User, error) {
	return f(ctx, loginName)
}

func TestLoginRejectsOversizedBody(t *testing.T) {
	var providerCalled bool
	handler := NewLogin(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		loginUserProviderFunc(func(context.Context, string) (User, error) {
			providerCalled = true
			return User{}, ErrUserNotFound
		}),
		jwtservice.NewTokenService("test-secret", time.Hour),
	)

	body := `{"login_name":"user","password":"` + strings.Repeat("x", maxLoginBodyBytes) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
	if providerCalled {
		t.Fatal("user provider was called for an oversized request")
	}
}

func TestLoginRejectsUnknownJSONField(t *testing.T) {
	handler := NewLogin(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		loginUserProviderFunc(func(context.Context, string) (User, error) {
			t.Fatal("user provider was called for invalid JSON")
			return User{}, nil
		}),
		jwtservice.NewTokenService("test-secret", time.Hour),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(
		`{"login_name":"user","password":"secret","unexpected":true}`,
	))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestLoginRateLimiterConcurrent(t *testing.T) {
	limiter := newLoginRateLimiter(time.Minute)
	var allowed atomic.Int32
	var wg sync.WaitGroup

	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Allow("192.0.2.1", "user") {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := allowed.Load(); got != maxAttemptsPerUser {
		t.Fatalf("allowed attempts = %d, want %d", got, maxAttemptsPerUser)
	}
}
