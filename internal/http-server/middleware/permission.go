package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type PermissionProvider interface {
	GetUserPermissions(
		ctx context.Context,
		userID int64,
	) ([]string, error)
}

func RequirePermission(
	log *slog.Logger,
	permissionProvider PermissionProvider,
	requiredPermission string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				writePermissionError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
				)
				return
			}

			permissions, err := permissionProvider.GetUserPermissions(
				r.Context(),
				claims.UserID,
			)
			if err != nil {
				log.Error(
					"failed to get user permissions",
					slog.Int64("user_id", claims.UserID),
					slog.String("permission", requiredPermission),
					slog.String("error", err.Error()),
				)

				writePermissionError(
					w,
					http.StatusInternalServerError,
					"internal server error",
				)
				return
			}

			if !hasPermission(permissions, requiredPermission) {
				log.Warn(
					"permission denied",
					slog.Int64("user_id", claims.UserID),
					slog.String("login_name", claims.LoginName),
					slog.String("permission", requiredPermission),
				)

				writePermissionError(
					w,
					http.StatusForbidden,
					"permission denied",
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func hasPermission(
	permissions []string,
	requiredPermission string,
) bool {
	for _, permission := range permissions {
		if permission == requiredPermission {
			return true
		}
	}

	return false
}

func writePermissionError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
