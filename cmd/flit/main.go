package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "flit: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 3 || os.Args[1] != "new" {
		fmt.Fprintln(os.Stderr, "usage: flit new MIGRATION-DIR")
		os.Exit(2)
	}

	dir := os.Args[2]
	di, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !di.IsDir() {
		return fmt.Errorf("stat %s: not a directory", dir)
	}

	name := fmt.Sprintf("%s-new-migration.sql", time.Now().Format("20060102150405"))
	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, nil, 0644); err != nil {
		return err
	}

	_, err = fmt.Println(path)
	return err
}
