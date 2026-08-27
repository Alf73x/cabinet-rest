package sqlite

import (
	"CabinetREST/internal/http-server/handlers/auth"
	"CabinetREST/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

/*
********************************************************************

	GetUserByLogin

********************************************************************
*/
func (s *Storage) GetUserByLogin(ctx context.Context, loginName string) (auth.User, error) {
	query := fmt.Sprintf(
		`SELECT %s, %s, %s, %s
		 FROM %s
		 WHERE %s = ?
		 LIMIT 1`,
		storage.Fld_common_id,
		storage.Fld_user_management_login_name,
		storage.Fld_user_management_user_password_hash,
		storage.Fld_user_management_enabled,
		storage.Tbl_user_management_users,
		storage.Fld_user_management_login_name,
	)

	var user auth.User
	var enabled int

	err := s.db.QueryRowContext(
		ctx,
		query,
		loginName,
	).Scan(
		&user.ID,
		&user.LoginName,
		&user.PasswordHash,
		&enabled,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, auth.ErrUserNotFound
	}

	if err != nil {
		return auth.User{}, fmt.Errorf(
			"get user by login: %w",
			err,
		)
	}
	user.Enabled = enabled != 0
	return user, nil
}
