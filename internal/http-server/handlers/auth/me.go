package auth

import (
	"log/slog"
	"net/http"

	appmiddleware "CabinetREST/internal/http-server/middleware"
)

func NewMe(log *slog.Logger, permissionProvider PermissionProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := appmiddleware.ClaimsFromContext(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		permissions, err := permissionProvider.GetUserPermissions(r.Context(), claims.UserID)
		if err != nil {
			log.Error("failed to get user permissions", slog.Int64("user_id", claims.UserID), slog.String("error", err.Error()))
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"user_id":     claims.UserID,
			"login_name":  claims.LoginName,
			"permissions": permissions,
		})
	}
}
