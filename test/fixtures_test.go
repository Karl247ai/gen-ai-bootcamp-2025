package test

import (
    "database/sql"
    "time"
)

// TestWord represents a test word fixture
type TestWord struct {
    Japanese string
    Romaji   string
    English  string
    CreatedAt time.Time
}

// Common test fixtures
var (
    testWords = []TestWord{
        {
            Japanese:  "猫",
            Romaji:    "neko",
            English:   "cat",
            CreatedAt: time.Now(),
        },
        {
            Japanese:  "犬",
            Romaji:    "inu",
            English:   "dog",
            CreatedAt: time.Now(),
        },
    }
)

// LoadTestWords loads test words into the database
func LoadTestWords(tx *sql.Tx) error {
    stmt, err := tx.Prepare(`
        INSERT INTO words (japanese, romaji, english, created_at)
        VALUES (?, ?, ?, ?)
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, word := range testWords {
        _, err = stmt.Exec(
            word.Japanese,
            word.Romaji,
            word.English,
            word.CreatedAt,
        )
        if err != nil {
            return err
        }
    }
    return nil
}