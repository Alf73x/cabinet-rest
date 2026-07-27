package sqlite

import (
	"CabinetREST/internal/storage"
	"context"
	"fmt"
)

func (s *Storage) GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	const op = "storage.sqlite.GetUserPermissions"

	query := fmt.Sprintf(`
			SELECT DISTINCT p.%s
			FROM %s AS gm
			JOIN %s AS ga ON ga.%s = gm.%s
			JOIN %s AS p ON p.%s = ga.%s
			WHERE gm.%s = ?
			ORDER BY p.%s `,
		storage.Fld_user_management_permission,
		storage.Tbl_user_management_group_members,
		storage.Tbl_user_management_group_access,
		storage.Fld_user_management_id_group,
		storage.Fld_user_management_id_group,
		storage.Tbl_user_management_permissions,
		storage.Fld_common_id,
		storage.Fld_user_management_id_permission,
		storage.Fld_user_management_id_user,
		storage.Fld_user_management_permission,
	)

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	permissions := make([]string, 16)
	for rows.Next() {
		var permission string

		if err := rows.Scan(&permission); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}

		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}

	return permissions, nil
}
