package flit

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"sync"
)

// A Migrator holds the configuration required to migrate a database.
// Call [New] to create a new Migrator.
type Migrator struct {
	db    *sql.DB
	fs    fs.FS
	glob  string
	guard GuardFunc
}

type migration struct {
	Sum  string // hex(sha256(Name))
	Name string
	SQL  string
}

// A ConfigOption can be passed to [New] to change the configuration.
// The [WithGlob] option configures the pattern used to load migration files.
// The [WithGuard] option configures the concurrency guard function.
type ConfigOption func(*Migrator)

// GuardFunc is called by [Migrator.Migrate] to manage concurrency.
type GuardFunc func(context.Context, *sql.Conn, func(context.Context, *sql.Conn) error) error

// New creates a new migrator for the given database, file system, and options.
func New(db *sql.DB, fsys fs.FS, options ...ConfigOption) *Migrator {
	m := &Migrator{
		db:    db,
		fs:    fsys,
		guard: new(mutexGuard).Guard,
		glob:  "*.sql",
	}

	for _, o := range options {
		o(m)
	}

	return m
}

// Migrate applies pending migrations to the database.
// It returns the names of the migrations that were applied.
//
// Migrations are loaded from .sql files in the root of the configured file system.
// The migrations are ordered by name before being applied.
// Each migration is executed as a single SQL statement.
// After a migration is completed a checksum of its name is recorded in the "flits" table,
// which is created automatically.
//
// Migrate is guarded by a mutex.
// This guard can be replaced by passing a [WithGuard] option to [New].
// For example, [GuardMySQL] uses MySQL's GET_LOCK and RELEASE_LOCK functions.
func (m *Migrator) Migrate(ctx context.Context) (applied []string, err error) {
	migrations, err := m.loadMigrations()
	if err != nil {
		return
	}

	conn, err := m.db.Conn(ctx)
	if err != nil {
		return
	}

	defer conn.Close()

	err = m.guard(ctx, conn, func(ctx context.Context, conn *sql.Conn) error {
		if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS flits (sum CHAR(64) PRIMARY KEY);`); err != nil {
			return fmt.Errorf("create flits table: %w", err)
		}

		completed, err := getCompletedMigrations(ctx, conn)
		if err != nil {
			return err
		}

		var pending []migration
		for sum, m := range migrations {
			if !slices.Contains(completed, sum) {
				pending = append(pending, m)
			}
		}

		// sort pending migrations by name
		slices.SortFunc(pending, func(a, b migration) int {
			return strings.Compare(a.Name, b.Name)
		})

		for _, m := range pending {
			if _, err := conn.ExecContext(ctx, m.SQL); err != nil {
				return fmt.Errorf("apply %s: %w", m.Name, err)
			}

			if _, err := conn.ExecContext(ctx, "INSERT INTO flits (sum) VALUES (?)", m.Sum); err != nil {
				return fmt.Errorf("record %s: %w", m.Name, err)
			}

			applied = append(applied, m.Name)
		}

		return nil
	})

	return
}

// loadMigrations reads every migration file matching the configured glob
// and returns a mapping keyed by the sha256 checksum of the file path.
func (m *Migrator) loadMigrations() (map[string]migration, error) {
	names, err := fs.Glob(m.fs, m.glob)
	if err != nil {
		return nil, err
	}

	stmts := make(map[string]migration)
	for _, name := range names {
		data, err := fs.ReadFile(m.fs, name)
		if err != nil {
			return nil, err
		}

		shasum := sha256.Sum256([]byte(name))
		hexsum := hex.EncodeToString(shasum[:])

		stmts[hexsum] = migration{
			Sum:  hexsum,
			Name: name,
			SQL:  string(data),
		}
	}

	return stmts, nil
}

// WithGlob configures Flit to load migration files matching the given glob.
func WithGlob(glob string) ConfigOption {
	return func(c *Migrator) {
		c.glob = glob
	}
}

// WithGuard configures Flit to call the given [GuardFunc] for concurrency control.
// For example, [GuardMySQL] uses MySQL's GET_LOCK and RELEASE_LOCK functions.
func WithGuard(g GuardFunc) ConfigOption {
	return func(c *Migrator) {
		c.guard = g
	}
}

// getCompletedMigrations loads the checksums of completed migrations from the flits table.
func getCompletedMigrations(ctx context.Context, conn *sql.Conn) (completed []string, err error) {
	rows, err := conn.QueryContext(ctx, "SELECT sum FROM flits")
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var sum string
		if err := rows.Scan(&sum); err != nil {
			return nil, err
		}

		completed = append(completed, sum)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return
}

// mutexGuard is the default guard.
type mutexGuard struct {
	sync.Mutex
}

func (g *mutexGuard) Guard(ctx context.Context, conn *sql.Conn, f func(context.Context, *sql.Conn) error) error {
	g.Lock()
	defer g.Unlock()
	return f(ctx, conn)
}
