package sqlite

import (
	"database/sql"
	"fmt"
	"testing"

	"CabinetREST/internal/storage"
)

func TestDbGetTeamMatchesReturnsTeamTreeError(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = (&Storage{db: db}).Db_GetTeamMatches(1, 1)
	if err == nil {
		t.Fatal("expected team tree error")
	}
}

func TestDbGetTeamMatchesReturnsEmptySliceForMissingTeam(t *testing.T) {
	db := newTeamMatchesTestDB(t)
	defer db.Close()

	matches, err := (&Storage{db: db}).Db_GetTeamMatches(1, 1)
	if err != nil {
		t.Fatalf("Db_GetTeamMatches() error = %v", err)
	}
	if matches == nil || len(matches) != 0 {
		t.Fatalf("Db_GetTeamMatches() = %#v, want empty non-nil slice", matches)
	}
}

func TestDbGetTeamMatchesReturnsEmptySliceForMissingSeason(t *testing.T) {
	db := newTeamMatchesTestDB(t)
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s) VALUES (1, 1, 0, NULL)",
		storage.Tbl_class_team,
		storage.Fld_common_id,
		storage.Fld_common_id_country,
		storage.Fld_common_private,
		storage.Fld_common_id_successor,
	)); err != nil {
		t.Fatal(err)
	}

	matches, err := (&Storage{db: db}).Db_GetTeamMatches(1, 999)
	if err != nil {
		t.Fatalf("Db_GetTeamMatches() error = %v", err)
	}
	if matches == nil || len(matches) != 0 {
		t.Fatalf("Db_GetTeamMatches() = %#v, want empty non-nil slice", matches)
	}
}

func newTeamMatchesTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	statements := []string{
		fmt.Sprintf("CREATE TABLE %s (%s INTEGER PRIMARY KEY, %s INTEGER, %s INTEGER)",
			storage.Tbl_countries,
			storage.Fld_common_id,
			storage.Fld_common_private,
			storage.Fld_countries_id_parent,
		),
		fmt.Sprintf("CREATE TABLE %s (%s INTEGER PRIMARY KEY, %s INTEGER, %s INTEGER, %s INTEGER)",
			storage.Tbl_class_team,
			storage.Fld_common_id,
			storage.Fld_common_id_country,
			storage.Fld_common_private,
			storage.Fld_common_id_successor,
		),
		fmt.Sprintf("CREATE TABLE %s (%s INTEGER PRIMARY KEY, %s TEXT, %s INTEGER)",
			storage.Tbl_class_season,
			storage.Fld_common_id,
			storage.Fld_class_season_options_1,
			storage.Fld_common_private,
		),
		fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES (1, 0)",
			storage.Tbl_countries,
			storage.Fld_common_id,
			storage.Fld_common_private,
		),
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			db.Close()
			t.Fatal(err)
		}
	}

	return db
}
