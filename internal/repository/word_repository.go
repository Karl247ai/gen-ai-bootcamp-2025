package repository

import (
    "database/sql"
    "context"
    "github.com/karl247ai/lang-portal/internal/models"
    "errors"
)

type WordRepository struct {
    db *sql.DB
}

func NewWordRepository(db *sql.DB) *WordRepository {
    return &WordRepository{db: db}
}

func (r *WordRepository) GetWords(ctx context.Context, limit, offset int) ([]models.Word, error) {
    query := `SELECT id, japanese, romaji, english, parts, created_at, updated_at 
              FROM words LIMIT ? OFFSET ?`
              
    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    if (err != nil) {
        return nil, err
    }
    defer rows.Close()

    var words []models.Word
    for rows.Next() {
        var w models.Word
        err := rows.Scan(&w.ID, &w.Japanese, &w.Romaji, &w.English, &w.Parts, &w.CreatedAt, &w.UpdatedAt)
        if err != nil {
            return nil, err
        }
        words = append(words, w)
    }
    return words, nil
}

func (r *WordRepository) CreateWord(ctx context.Context, word *models.Word) error {
    query := `
        INSERT INTO words (japanese, romaji, english, parts, created_at, updated_at)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `
    
    result, err := r.db.ExecContext(ctx, query, 
        word.Japanese, 
        word.Romaji, 
        word.English, 
        word.Parts,
    )
    if err != nil {
        return err
    }

    id, err := result.LastInsertId()
    if err != nil {
        return err
    }

    word.ID = id
    return nil
}

func (r *WordRepository) UpdateWord(ctx context.Context, id int64, word *models.Word) error {
    query := `
        UPDATE words 
        SET japanese = ?, romaji = ?, english = ?, parts = ?, updated_at = CURRENT_TIMESTAMP
        WHERE id = ?
    `
    
    result, err := r.db.ExecContext(ctx, query, 
        word.Japanese, 
        word.Romaji, 
        word.English, 
        word.Parts,
        id,
    )
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("word not found")
    }

    return nil
}

func (r *WordRepository) DeleteWord(ctx context.Context, id int64) error {
    result, err := r.db.ExecContext(ctx, "DELETE FROM words WHERE id = ?", id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("word not found")
    }

    return nil
}

func (r *WordRepository) GetWordsCount(ctx context.Context) (int64, error) {
    var count int64
    err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM words").Scan(&count)
    return count, err
}