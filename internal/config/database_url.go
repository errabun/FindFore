package config

import (
	"fmt"
	"net/url"
	"os"
)

// ResolveDatabaseURL returns a Postgres connection string from either DATABASE_URL
// or Cloud SQL component env vars used on Cloud Run.
func ResolveDatabaseURL() (string, error) {
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL, nil
	}

	instance := os.Getenv("INSTANCE_CONNECTION_NAME")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")

	if instance == "" || user == "" || pass == "" || name == "" {
		return "", fmt.Errorf("DATABASE_URL is required, or set INSTANCE_CONNECTION_NAME, DB_USER, DB_PASS, and DB_NAME for Cloud SQL")
	}

	return fmt.Sprintf(
		"postgres://%s:%s@/%s?host=/cloudsql/%s",
		url.QueryEscape(user),
		url.QueryEscape(pass),
		name,
		instance,
	), nil
}
