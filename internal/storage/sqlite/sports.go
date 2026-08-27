package sqlite

/*********************************************************************
  Db_GetSports
**********************************************************************/import (
	"CabinetREST/internal/storage"
	"fmt"
)

func (s *Storage) Db_GetSports() ([]storage.TblSport, error) {
	const _FunctionName = "storage.sqlite.Db_GetSports"

	sSQL := fmt.Sprintf(`SELECT %s, %s, %s  
		FROM %s  
        WHERE %s=1 AND IFNULL(%s, '') NOT LIKE '%%Type=V%%' COLLATE NOCASE
		ORDER BY %s`, storage.Fld_common_id, storage.Fld_common_name, storage.Fld_class_base_options,
		storage.Tbl_class_base,
		storage.Fld_class_base_available, storage.Fld_class_base_options, storage.Fld_common_id)
	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}

	defer rows.Close()

	var sports []storage.TblSport
	for rows.Next() {
		var sps storage.TblSport
		err := rows.Scan(
			&sps.ID,
			&sps.Name,
			&sps.BaseOption,
		)
		if err != nil {
			return nil, err
		}
		sports = append(sports, sps)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sports, nil
}
