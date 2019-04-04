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

func issue1() {
	readDB, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/mysql?parseTime=true&charset=utf8mb4")
	if err != nil {
		panic(err.Error())
	}

	rows, err := readDB.Query("SELECT 1")
	if err != nil {
		panic(err)
	}
	defer rows.Close() // OK
}

func issue3() {
	readDB, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/mysql?parseTime=true&charset=utf8mb4")
	if err != nil {
		panic(err.Error())
	}

	rows, err := readDB.Query("SELECT 1") // want "rows.Close must be called in defer function"
	rows.Close()
}

func skip() {
	fmt.Print("skip")
}
