package db_test

import (
	"path/filepath"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/db"
)

func TestOpenMigratesSchema(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	for _, table := range []string{"users", "invites", "sessions"} {
		var name string
		err := conn.QueryRow(
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table,
		).Scan(&name)
		if err != nil {
			t.Errorf("таблица %q не создана миграциями: %v", table, err)
		}
	}
}

func TestOpenSetsPragmas(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	var journalMode string
	if err := conn.QueryRow(`PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want wal", journalMode)
	}

	var fk int
	if err := conn.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}
}

func TestOpenEnforcesForeignKeys(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	_, err = conn.Exec(
		`INSERT INTO invites (code, created_by, created_at) VALUES ('x', 999, '2026-01-01T00:00:00Z')`,
	)
	if err == nil {
		t.Fatal("вставка с несуществующим created_by прошла — foreign_keys не работают")
	}
}

func TestOpenIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	for i := 0; i < 2; i++ {
		conn, err := db.Open(path)
		if err != nil {
			t.Fatalf("Open #%d: %v", i+1, err)
		}
		conn.Close()
	}
}
