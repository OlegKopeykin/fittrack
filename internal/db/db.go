// Package db — открытие SQLite с боевыми PRAGMA и запуск embedded-миграций.
package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Open открывает SQLite-базу по пути path (создавая файл при необходимости),
// включает WAL, foreign_keys и busy_timeout и накатывает миграции goose.
// Пул ограничен одной коннекцией: единственный писатель — норма для SQLite
// и дружелюбно к потоковой репликации.
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)",
		path,
	)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("открытие %s: %w", path, err)
	}
	conn.SetMaxOpenConns(1)

	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("sqlite3"); err != nil {
		conn.Close()
		return nil, err
	}
	if err := goose.Up(conn, "migrations"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("миграции: %w", err)
	}
	return conn, nil
}
