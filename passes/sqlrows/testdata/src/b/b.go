package b

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

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			panic(err)
		}
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}
}

func badQueryContext() {
	var ctx context.Context
	var db *sql.DB

	rows, err := db.QueryContext(ctx, "SELECT * FROM users") // want "rows.Err must be called"
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
}

func skip() {
	fmt.Print("skip")
}
