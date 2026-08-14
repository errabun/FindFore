package main

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMigrationsFromEmptyDatabase(t *testing.T) {
	dsn := os.Getenv("FINDFORE_MIGRATE_TEST_URL")
	if dsn == "" {
		t.Skip("set FINDFORE_MIGRATE_TEST_URL to run migration smoke test against an empty database")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.Ping())

	_, err = db.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	require.NoError(t, err)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err)

	root := findRepoRoot(t)
	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.Join(root, "migrations"), "postgres", driver)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err)
	}

	var groupIDCol string
	err = db.QueryRow(`
		SELECT pg_get_constraintdef(c.oid)
		FROM pg_constraint c
		JOIN pg_class t ON c.conrelid = t.oid
		WHERE t.relname = 'events' AND c.contype = 'f'
		  AND pg_get_constraintdef(c.oid) LIKE '%group_id%'
	`).Scan(&groupIDCol)
	require.NoError(t, err)
	require.Contains(t, groupIDCol, "ON DELETE SET NULL")
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "migrations")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("migrations directory not found")
	return ""
}
