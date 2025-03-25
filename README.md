# Flit

Flit is a Go package that implements SQL database migrations.
It is built on top of Go's standard [`database/sql`](https://pkg.go.dev/database/sql) package.
There are many packages like this one and all of them are better, but Flit is small and easy to understand.

Flit reads migrations from `.sql` files and executes each one as a single SQL statement.
Completed migrations are recorded in the `flits` table, which is created automatically.

To use Flit, create a new migrator and call `Migrate` when your process starts.
You can handle concurrent processes by configuring a guard function like the following example.

## Example

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/180-studios/flit"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		panic(err)
	}

	defer db.Close()

	dir := os.Args[1]

	// read migrations from dir/*.sql (or embed them)
	// manage concurrency with MySQL's GET_LOCK/RELEASE_LOCK
	m := flit.New(db, os.DirFS(dir), flit.WithGuard(flit.GuardMySQL))
	applied, err := m.Migrate(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(applied)
}
```

## Development

The MySQL tests are skipped unless the `TEST_MYSQL_DSN` environment variable is set.

## Dependencies

Flit doesn't have any runtime dependencies other than the standard library.
The tests and examples depend on the `github.com/go-sql-driver/mysql` and `github.com/mattn/go-sqlite3` modules.
