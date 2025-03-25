package flit

import (
	"context"
	"database/sql"
	"errors"
)

// GuardMySQL manages migration concurrency with MySQL's GET_LOCK and RELEASE_LOCK functions.
// It gets a lock named "flit" before calling f and releases it after f returns.
// GuardMySQL blocks until the lock is acquired or ctx is done.
// Use this guard function by passing a [WithGuard] option to [New].
func GuardMySQL(ctx context.Context, conn *sql.Conn, f func(context.Context, *sql.Conn) error) (err error) {
	if _, err := conn.ExecContext(ctx, "SELECT GET_LOCK('flit', -1)"); err != nil {
		return err
	}

	defer func() {
		_, re := conn.ExecContext(ctx, "SELECT RELEASE_LOCK('flit')")
		err = errors.Join(err, re)
	}()

	return f(ctx, conn)
}
