package middleware

import (
	"log/slog"
	"net/http"
)

type PermissionStorage interface {
	HasPermission(
		userID int64,
		permission string,
	) (bool, error)
}

/*
RequirePermission проверяет, есть ли у текущего пользователя необходимое право.

Важно: перед RequirePermission обязательно должен выполняться JWT middleware.

JWT:
  - проверяет токен;
  - получает claims;
  - сохраняет claims в context.

RequirePermission:
  - получает claims из context;
  - берёт claims.UserID;
  - проверяет право пользователя в БД;
  - разрешает или запрещает выполнение handler.
*/
func RequirePermission(log *slog.Logger, storage PermissionStorage, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok || claims == nil {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			hasPermission, err := storage.HasPermission(claims.UserID, permission)
			if err != nil {
				log.Error("failed to check permission", slog.Int64("user_id", claims.UserID), slog.String("permission", permission), slog.String("error", err.Error()))
				writeJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			if !hasPermission {
				writeJSONError(w, http.StatusForbidden, "permission denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
