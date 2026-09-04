package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"CabinetREST/internal/storage"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: open database: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%s: ping database: %w", op, err)
	}

	// Проверяем таблицы, без которых приложение работать не может.
	requiredTables := []string{
		storage.Tbl_class_season,
		storage.Tbl_class_team,
		storage.Tbl_countries,
		storage.Tbl_sport_results,
		storage.Tbl_sport_tables,
	}

	for _, table := range requiredTables {
		var exists int

		err := db.QueryRowContext(
			ctx,
			`SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ? )`,
			table,
		).Scan(&exists)

		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("%s: check table %q: %w", op, table, err)
		}

		if exists == 0 {
			_ = db.Close()
			return nil, fmt.Errorf("%s: required table %q is missing", op, table)
		}
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	var result int
	if err := s.db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("query database: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected database response")
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
