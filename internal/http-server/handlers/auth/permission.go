package auth

import "context"

type PermissionProvider interface {
	GetUserPermissions(
		ctx context.Context,
		userID int64,
	) ([]string, error)
}
