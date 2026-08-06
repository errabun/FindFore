package config

import (
	"testing"
)

func TestResolveDatabaseURL(t *testing.T) {
	t.Run("uses DATABASE_URL when set", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://localhost:5432/findfore?sslmode=disable")
		t.Setenv("INSTANCE_CONNECTION_NAME", "")
		t.Setenv("DB_USER", "")
		t.Setenv("DB_PASS", "")
		t.Setenv("DB_NAME", "")

		got, err := ResolveDatabaseURL()
		if err != nil {
			t.Fatalf("ResolveDatabaseURL() error = %v", err)
		}
		if got != "postgres://localhost:5432/findfore?sslmode=disable" {
			t.Fatalf("ResolveDatabaseURL() = %q", got)
		}
	})

	t.Run("builds Cloud SQL unix socket URL from components", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("INSTANCE_CONNECTION_NAME", "findfore-prod:us-central1:findfore-db")
		t.Setenv("DB_USER", "findfore")
		t.Setenv("DB_PASS", "p@ss:word")
		t.Setenv("DB_NAME", "findfore")

		got, err := ResolveDatabaseURL()
		if err != nil {
			t.Fatalf("ResolveDatabaseURL() error = %v", err)
		}

		want := "postgres://findfore:p%40ss%3Aword@/findfore?host=/cloudsql/findfore-prod:us-central1:findfore-db"
		if got != want {
			t.Fatalf("ResolveDatabaseURL() = %q, want %q", got, want)
		}
	})

	t.Run("returns error when no database config is provided", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("INSTANCE_CONNECTION_NAME", "")
		t.Setenv("DB_USER", "")
		t.Setenv("DB_PASS", "")
		t.Setenv("DB_NAME", "")

		if _, err := ResolveDatabaseURL(); err == nil {
			t.Fatal("expected error when database config is missing")
		}
	})
}
