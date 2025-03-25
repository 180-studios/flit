package flit_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/180-studios/flit"
	"github.com/180-studios/flit/mysqltest"
	"github.com/180-studios/flit/sqlitetest"
	"github.com/google/go-cmp/cmp"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func ExampleMigrator() {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	m := flit.New(db, os.DirFS("testdata/example"))
	applied, err := m.Migrate(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(applied)
	// Output: [001-first.sql 002-second.sql]
}

func TestMySQL(t *testing.T) {
	dsn, ok := os.LookupEnv("TEST_MYSQL_DSN")
	if !ok {
		t.Skip("TEST_MYSQL_DSN is not set")
	}

	db := mysqltest.NewDB(t, dsn)
	m := flit.New(db, os.DirFS("testdata/example"), flit.WithGuard(flit.GuardMySQL))
	applied, err := m.Migrate(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff([]string{"001-first.sql", "002-second.sql"}, applied); diff != "" {
		t.Errorf("applied migrations differ (-want +got):\n%s", diff)
	}
}

func TestOrder(t *testing.T) {
	db := sqlitetest.NewDB(t)
	m := flit.New(db, os.DirFS("testdata/lexical-order"))
	applied, err := m.Migrate(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	expect := []string{
		"002-first.sql",
		"01-second.sql",
	}

	if diff := cmp.Diff(expect, applied); diff != "" {
		t.Errorf("applied migrations differ (-want +got):\n%s", diff)
	}
}

func TestMultipleRuns(t *testing.T) {
	db := sqlitetest.NewDB(t)
	m := flit.New(db, os.DirFS("testdata/multiple-runs/first"))
	applied, err := m.Migrate(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	expect := []string{
		"001-first.sql",
	}

	if diff := cmp.Diff(expect, applied); diff != "" {
		t.Errorf("first run: applied migrations differ (-want +got):\n%s", diff)
	}

	// second is a superset of first
	m = flit.New(db, os.DirFS("testdata/multiple-runs/second"))
	applied, err = m.Migrate(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	expect = []string{
		"002-second.sql",
	}

	if diff := cmp.Diff(expect, applied); diff != "" {
		t.Errorf("second run: applied migrations differ (-want +got):\n%s", diff)
	}

	applied, err = m.Migrate(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if len(applied) != 0 {
		t.Errorf("third run: expected no migrations, got %v", applied)
	}
}

func TestWithGlob(t *testing.T) {
	db := sqlitetest.NewDB(t)
	m := flit.New(db, os.DirFS("testdata"), flit.WithGlob("example/*.sql"))
	applied, err := m.Migrate(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	expect := []string{
		"example/001-first.sql",
		"example/002-second.sql",
	}

	if diff := cmp.Diff(expect, applied); diff != "" {
		t.Errorf("applied migrations differ (-want +got):\n%s", diff)
	}
}
