package testutil

import (
	"database/sql"
	"testing"

	"github.com/your-org/lang-portal/internal/database"
)

// NewTestDB creates a new test database and returns the connection
func NewTestDB(t testing.TB) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// ExecuteSQL executes the SQL statements in the given file
func ExecuteSQL(t testing.TB, db *sql.DB, path string) {
	if err := database.ExecuteFile(db, path); err != nil {
		t.Fatalf("Failed to execute SQL file %s: %v", path, err)
	}
}

func CleanupDB(t testing.TB, db *sql.DB) {
	tables := []string{"words", "groups", "words_groups"}
	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table)
		if err != nil {
			t.Fatalf("Failed to clean up table %s: %v", table, err)
		}
	}
} 