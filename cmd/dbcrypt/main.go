/*
Из корня CabinetREST:

go build -ldflags="-s -w" -o dbcrypt.exe ./cmd/dbcrypt
// go build -o dbcrypt.exe ./cmd/dbcrypt - с debug информацией

dbcrypt.exe "C:\Cabinet\CabinetREST\storage\cabinet.db" "C:\Cabinet\CabinetREST\storage\ecabinet.db" ""
*/
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage:")
		fmt.Println(`  dbcrypt.exe <source.db> <destination.db> <password>`)
		os.Exit(1)
	}

	source := os.Args[1]
	destination := os.Args[2]
	password := os.Args[3]

	if err := encryptDatabase(source, destination, password); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Database encrypted successfully:")
	fmt.Println(destination)
}

func encryptDatabase(source, destination, password string) error {
	if source == destination {
		return fmt.Errorf("source and destination must be different")
	}

	if password == "" {
		return fmt.Errorf("password is empty")
	}

	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("source database: %w", err)
	}

	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("destination file already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("destination file: %w", err)
	}

	db, err := sql.Open("sqlite3", source)
	if err != nil {
		return fmt.Errorf("open source database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("get database connection: %w", err)
	}
	defer conn.Close()

	var cipherVersion string

	if err := conn.QueryRowContext(
		ctx,
		"PRAGMA cipher_version",
	).Scan(&cipherVersion); err != nil {
		return fmt.Errorf("SQLCipher is not available: %w", err)
	}

	fmt.Println("SQLCipher:", cipherVersion)

	var sourceObjects int

	if err := conn.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM sqlite_master",
	).Scan(&sourceObjects); err != nil {
		return fmt.Errorf("read source database: %w", err)
	}

	fmt.Printf("Source schema objects: %d\n", sourceObjects)

	escapedDestination := strings.ReplaceAll(destination, "'", "''")
	escapedPassword := strings.ReplaceAll(password, "'", "''")

	attachSQL := fmt.Sprintf(
		"ATTACH DATABASE '%s' AS encrypted KEY '%s'",
		escapedDestination,
		escapedPassword,
	)

	if _, err := conn.ExecContext(ctx, attachSQL); err != nil {
		os.Remove(destination)
		return fmt.Errorf("attach encrypted database: %w", err)
	}

	if _, err := conn.ExecContext(
		ctx,
		"SELECT sqlcipher_export('encrypted')",
	); err != nil {
		conn.ExecContext(ctx, "DETACH DATABASE encrypted")
		os.Remove(destination)
		return fmt.Errorf("export database: %w", err)
	}

	if _, err := conn.ExecContext(
		ctx,
		"DETACH DATABASE encrypted",
	); err != nil {
		os.Remove(destination)
		return fmt.Errorf("detach encrypted database: %w", err)
	}

	if err := verifyDatabase(destination, password); err != nil {
		os.Remove(destination)
		return err
	}

	return nil
}

func verifyDatabase(filename, password string) error {
	key := url.QueryEscape(password)

	dsn := fmt.Sprintf(
		"%s?_pragma_key=%s",
		filename,
		key,
	)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("verify: open database: %w", err)
	}
	defer db.Close()

	var cipherVersion string

	if err := db.QueryRow(
		"PRAGMA cipher_version",
	).Scan(&cipherVersion); err != nil {
		return fmt.Errorf("verify: SQLCipher not available: %w", err)
	}

	var count int

	if err := db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master",
	).Scan(&count); err != nil {
		return fmt.Errorf("verify encrypted database: %w", err)
	}

	fmt.Printf("Verification OK, schema objects: %d\n", count)

	return nil
}
