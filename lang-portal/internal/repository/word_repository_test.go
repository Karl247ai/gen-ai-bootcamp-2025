package repository

import (
    "testing"
    "context"
    "database/sql"
    "github.com/stretchr/testify/assert"
    _ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    assert.NoError(t, err)

    // Create test table
    _, err = db.Exec(`
        CREATE TABLE words (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            japanese TEXT NOT NULL,
            romaji TEXT NOT NULL,
            english TEXT NOT NULL,
            parts JSON,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    assert.NoError(t, err)
    return db
}

func TestWordRepository_GetWords(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Insert test data
    _, err := db.Exec(`
        INSERT INTO words (japanese, romaji, english) 
        VALUES ('こんにちは', 'konnichiwa', 'hello')
    `)
    assert.NoError(t, err)

    repo := NewWordRepository(db)
    words, err := repo.GetWords(context.Background(), 10, 0)
    
    assert.NoError(t, err)
    assert.Len(t, words, 1)
    assert.Equal(t, "こんにちは", words[0].Japanese)
    assert.Equal(t, "konnichiwa", words[0].Romaji)
    assert.Equal(t, "hello", words[0].English)
}