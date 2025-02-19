package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/your-org/lang-portal/internal/database"
	"github.com/your-org/lang-portal/internal/errors"
	"github.com/your-org/lang-portal/internal/models"
)

const defaultTimeout = 5 * time.Second

type WordRepository struct {
	db *sql.DB
}

func NewWordRepository(db *sql.DB) *WordRepository {
	return &WordRepository{db: db}
}

func (r *WordRepository) Create(ctx context.Context, word *models.Word) error {
	query := `
		INSERT INTO words (japanese, romaji, english, parts, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	word.CreatedAt = now
	word.UpdatedAt = now

	result, err := database.ExecContext(ctx, r.db, defaultTimeout, query,
		word.Japanese,
		word.Romaji,
		word.English,
		word.Parts,
		word.CreatedAt,
		word.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(err, errors.ErrDBQuery, "failed to create word")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, errors.ErrDBQuery, "failed to get last insert ID")
	}

	word.ID = int(id)
	return nil
}

func (r *WordRepository) GetByID(ctx context.Context, id int) (*models.Word, error) {
	query := `
		SELECT id, japanese, romaji, english, parts, created_at, updated_at
		FROM words WHERE id = ?
	`

	rows, err := database.QueryContext(ctx, r.db, defaultTimeout, query, id)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDBQuery, "failed to query word")
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New(errors.ErrDBNotFound, "word not found")
	}

	word := &models.Word{}
	err = rows.Scan(
		&word.ID,
		&word.Japanese,
		&word.Romaji,
		&word.English,
		&word.Parts,
		&word.CreatedAt,
		&word.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDBQuery, "failed to scan word")
	}

	return word, nil
}

func (r *WordRepository) List(ctx context.Context, offset, limit int) ([]*models.Word, error) {
	query := `
		SELECT id, japanese, romaji, english, parts, created_at, updated_at
		FROM words ORDER BY id LIMIT ? OFFSET ?
	`

	rows, err := database.QueryContext(ctx, r.db, defaultTimeout, query, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDBQuery, "failed to list words")
	}
	defer rows.Close()

	var words []*models.Word
	for rows.Next() {
		word := &models.Word{}
		err = rows.Scan(
			&word.ID,
			&word.Japanese,
			&word.Romaji,
			&word.English,
			&word.Parts,
			&word.CreatedAt,
			&word.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDBQuery, "failed to scan word")
		}
		words = append(words, word)
	}

	return words, nil
} 