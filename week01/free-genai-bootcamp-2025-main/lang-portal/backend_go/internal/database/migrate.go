package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migration represents a database migration
type Migration struct {
	ID      int
	Name    string
	Content string
}

// Migrate runs all pending migrations
func Migrate(db *sql.DB, migrationsPath string) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("error getting applied migrations: %w", err)
	}

	// Get available migrations
	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return fmt.Errorf("error loading migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if !isApplied(migration, applied) {
			if err := runMigration(db, migration); err != nil {
				return fmt.Errorf("error running migration %s: %w", migration.Name, err)
			}
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM migrations ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrations = append(migrations, name)
	}
	return migrations, nil
}

func loadMigrations(dir string) ([]Migration, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			content, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}

			migrations = append(migrations, Migration{
				Name:    file.Name(),
				Content: string(content),
			})
		}
	}

	// Sort migrations by filename
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}

func isApplied(migration Migration, applied []string) bool {
	for _, name := range applied {
		if name == migration.Name {
			return true
		}
	}
	return false
}

func runMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Run migration
	if _, err := tx.Exec(migration.Content); err != nil {
		return err
	}

	// Record migration
	if _, err := tx.Exec("INSERT INTO migrations (name) VALUES (?)", migration.Name); err != nil {
		return err
	}

	return tx.Commit()
} 