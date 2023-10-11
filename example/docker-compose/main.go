package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

type users struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	ctx := context.Background()

	connStr, err := os.ReadFile("/run/secrets/connection_string")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error read connection secret: %v\n", err)
		os.Exit(1)
	}

	conn, err := pgx.Connect(ctx, string(connStr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	http.ListenAndServe("0.0.0.0:80", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			rows, err := conn.Query(ctx, "select * from users")
			if err != nil {
				fmt.Printf("error db query: %s", err)
				return
			}

			users, err := pgx.CollectRows(rows, pgx.RowToStructByName[users])
			if err != nil {
				fmt.Printf("error collect rows: %s", err)
				return
			}

			jsonUsers, err := json.Marshal(users)
			if err != nil {
				fmt.Printf("error marshal json: %s", err)
			}

			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Write(jsonUsers)
		}),
	)
}
