package sqlite

import (
	"CabinetREST/internal/storage"
	"fmt"
	"strings"
)

/*********************************************************************
  Territories for parent.
*********************************************************************/

func (s *Storage) Db_GetTerritories(parentID int) ([]storage.TblCountry, error) {
	const _FunctionName = "storage.sqlite.Db_GetCountries"
	if parentID > 0 {
		var parentIsPublic bool
		query := "SELECT EXISTS (SELECT 1 FROM " + storage.Tbl_countries +
			" WHERE " + storage.Fld_common_id + "=? AND IFNULL(" + storage.Fld_common_private + ", 0) <> 1)"
		if err := s.db.QueryRow(query, parentID).Scan(&parentIsPublic); err != nil {
			return nil, fmt.Errorf("%s: check parent: %w", _FunctionName, err)
		}
		if !parentIsPublic {
			return []storage.TblCountry{}, nil
		}
	}

	if parentID <= 0 {
		parentID = -1
	}

	sSQL := fmt.Sprintf(`SELECT c.%s, c.%s, 
	    CASE WHEN EXISTS (SELECT 1 FROM %s c2 WHERE c2.%s = c.%s AND IFNULL(c2.%s, 0) <> 1) THEN 1
    	ELSE 0 END AS has_children,
		IFNULL(%s,"")  /* sort_order */
		FROM %s c 
		WHERE IFNULL(%s, 0) <> 1 AND IFNULL(c.%s, -1)=%d 
		ORDER BY %s, %s`, storage.Fld_common_id, storage.Fld_common_name, storage.Tbl_countries, storage.Fld_countries_id_parent, storage.Fld_common_id,
		storage.Fld_common_private, storage.Fld_common_sort_order,
		storage.Tbl_countries,
		storage.Fld_common_private, storage.Fld_countries_id_parent,
		parentID, storage.Fld_countries_sort_order, storage.Fld_common_name)

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}
	defer rows.Close()

	var countries []storage.TblCountry
	for rows.Next() {
		var cntr storage.TblCountry
		err := rows.Scan(
			&cntr.ID,
			&cntr.Name,
			&cntr.HasChildren,
			&cntr.SortOrder,
		)
		if err != nil {
			return nil, err
		}
		countries = append(countries, cntr)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}

/*********************************************************************
  Territories. Search for text
*********************************************************************/

func (s *Storage) Db_SearchTerritories(filter string) ([]storage.TblCountry, error) {
	const _FunctionName = "storage.sqlite.Db_SearchTerritories"

	sSQL := fmt.Sprintf(`SELECT c.%s, c.%s, 
	    0 AS has_children,
		IFNULL(%s,"")  /* sort_order */
		FROM %s c 
		WHERE IFNULL(%s, 0) <> 1
		ORDER BY %s, %s`, storage.Fld_common_id, storage.Fld_common_name,
		storage.Fld_common_sort_order,
		storage.Tbl_countries,
		storage.Fld_common_private,
		storage.Fld_countries_sort_order, storage.Fld_common_name)

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}
	defer rows.Close()

	var countries []storage.TblCountry
	filterLower := strings.ToLower(filter)
	for rows.Next() {
		var cntr storage.TblCountry
		err := rows.Scan(
			&cntr.ID,
			&cntr.Name,
			&cntr.HasChildren,
			&cntr.SortOrder,
		)
		if err != nil {
			return nil, err
		}

		if strings.Contains(strings.ToLower(cntr.Name), filterLower) {
			countries = append(countries, cntr)
		}

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}

/*********************************************************************
  Db_PathTerritories
**********************************************************************/

func (s *Storage) Db_PathTerritories(id int) ([]int, error) {
	const _FunctionName = "storage.sqlite.Db_PathTerritories"

	sSQL := fmt.Sprintf(`
	WITH RECURSIVE parents AS (
		SELECT %s, %s, 0 AS level FROM %s WHERE id=%v AND IFNULL(%s, 0) <> 1
		UNION ALL
		SELECT t.%s, t.%s, p.level + 1 FROM %s t
		JOIN parents p ON t.%s = p.%s
		WHERE IFNULL(t.%s, 0) <> 1
	)
	SELECT %s FROM parents ORDER BY level DESC`,
		storage.Fld_common_id, storage.Fld_countries_id_parent, storage.Tbl_countries, id, storage.Fld_common_private,
		storage.Fld_common_id, storage.Fld_countries_id_parent, storage.Tbl_countries,
		storage.Fld_common_id, storage.Fld_countries_id_parent, storage.Fld_common_private,
		storage.Fld_common_id,
	)

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}
	defer rows.Close()

	var path []int
	for rows.Next() {
		var cid int
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		path = append(path, cid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return path, nil
}
