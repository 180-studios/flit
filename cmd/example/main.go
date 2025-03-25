// Example is the example code embedded in the README.
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
