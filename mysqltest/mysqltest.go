// Package mysqltest provides a helper to create test-scoped MySQL databases.
package mysqltest

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"io"
	"testing"

	"github.com/go-sql-driver/mysql"
)

// NewDB creates a new MySQL database that is dropped after the test.
// It connects to the database described by templateDSN to execute CREATE DATABASE and DROP DATABASE statements.
// The new database is named by adding a random suffix to the database name in templateDSN.
func NewDB(t *testing.T, templateDSN string) *sql.DB {
	t.Helper()

	template, err := mysql.ParseDSN(templateDSN)
	if err != nil {
		t.Fatal(err)
	}

	root, err := sql.Open("mysql", template.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Errorf("close %s: %v", template.FormatDSN(), err)
		}
	})

	randomBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		t.Fatal(err)
	}

	dsn := *template
	dsn.DBName = "mysqltest_" + hex.EncodeToString(randomBytes)

	if template.DBName != "" {
		dsn.DBName = template.DBName + "_" + dsn.DBName
	}

	if _, err := root.Exec("CREATE DATABASE " + dsn.DBName); err != nil {
		t.Fatalf("create %s: %v", dsn.DBName, err)
	}

	t.Cleanup(func() {
		if _, err := root.Exec("DROP DATABASE " + dsn.DBName); err != nil {
			t.Errorf("drop %s: %v", dsn.DBName, err)
		}
	})

	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close %s: %v", dsn.FormatDSN(), err)
		}
	})

	return db
}
