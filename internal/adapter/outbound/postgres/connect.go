package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	// Heroku Postgres requires SSL but doesn't include sslmode in the URL
	if !strings.Contains(databaseURL, "sslmode=") && !strings.Contains(databaseURL, "localhost") {
		if strings.Contains(databaseURL, "?") {
			databaseURL += "&sslmode=require"
		} else {
			databaseURL += "?sslmode=require"
		}
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}
