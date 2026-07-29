package sqlite

import (
	"CabinetREST/internal/storage"
	"fmt"
)

/*
HasPermission проверяет, есть ли у пользователя указанное право.

Метод возвращает:

	true, nil		право найдено;
	false, nil		право не найдено;
	false, error	произошла ошибка при обращении к БД.
*/
func (s *Storage) HasPermission(userID int64, permission string) (bool, error) {
	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM %s gm
			INNER JOIN %s ga ON ga.%s = gm.%s
			INNER JOIN %s p	ON p.%s = ga.%s
			WHERE gm.%s = ? AND p.%s = ?
		)`,
		storage.Tbl_user_management_group_members,
		storage.Tbl_user_management_group_access,
		storage.Fld_user_management_id_group,
		storage.Fld_user_management_id_group,
		storage.Tbl_user_management_permissions,
		storage.Fld_user_management_id_permission,
		storage.Fld_user_management_id_permission,
		storage.Fld_user_management_id_user,
		storage.Fld_user_management_permission,
	)

	var hasPermission bool

	err := s.db.QueryRow(query, userID, permission).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("check permission %q for user %d: %w", permission, userID, err)
	}

	return hasPermission, nil
}
