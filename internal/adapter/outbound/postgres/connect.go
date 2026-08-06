package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	// Cloud SQL on Cloud Run uses a Unix socket and does not need sslmode.
	// Remote TCP connections without sslmode still default to require.
	if !strings.Contains(databaseURL, "sslmode=") &&
		!strings.Contains(databaseURL, "localhost") &&
		!strings.Contains(databaseURL, "/cloudsql/") {
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
