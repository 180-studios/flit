// Package sqlitetest provides a test helper to create an in-memory SQLite database.
// It uses the github.com/mattn/go-sqlite3 driver.
package sqlitetest

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("sqlitetest: open database: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("sqlitetest: open database: %v", err)
		}
	})

	return db
}
