package a

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

func goodQueryContext() {
	var ctx context.Context
	var db *sql.DB
	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
}

func badQueryContext() {
	var ctx context.Context
	var db *sql.DB

	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
	defer rows.Close() // want "using rows before checking for errors"
	if err != nil {
		log.Fatal(err)
	}
}

func closeNotCalled() {
	var ctx context.Context
	var db *sql.DB

	rows, err := db.QueryContext(ctx, "SELECT * FROM users") // want "rows.Close must be called"
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(rows)

	_, err = db.QueryContext(ctx, "SELECT * FROM users") // want "rows.Close must be called"
	if err != nil {
		log.Fatal(err)
	}
}

func skip() {
	fmt.Print("skip")
}
