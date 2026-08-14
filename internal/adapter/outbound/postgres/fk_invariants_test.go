package postgres

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupForeignKeyInvariants(t *testing.T) {
	root := repoRoot(t)
	groupsSQL := readMigration(t, root, "000014_groups.up.sql")
	postsSQL := readMigration(t, root, "000015_group_posts.up.sql")
	eventsSQL := readMigration(t, root, "000016_group_events.up.sql")

	require.Contains(t, groupsSQL, "REFERENCES groups(id) ON DELETE CASCADE")
	require.GreaterOrEqual(t, strings.Count(groupsSQL, "ON DELETE CASCADE"), 2)
	require.Contains(t, postsSQL, "REFERENCES groups(id) ON DELETE CASCADE")
	require.Contains(t, eventsSQL, "REFERENCES groups(id) ON DELETE SET NULL")
	require.NotContains(t, eventsSQL, "ON DELETE CASCADE")
}

func repoRoot(t *testing.T) string {
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

func readMigration(t *testing.T, root, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "migrations", name))
	require.NoError(t, err)
	return string(b)
}
