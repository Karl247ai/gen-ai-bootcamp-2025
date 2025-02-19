package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/your-org/lang-portal/internal/database"
	"github.com/your-org/lang-portal/internal/errors"
	"github.com/your-org/lang-portal/internal/models"
)

type WordGroupRepository struct {
	db *sql.DB
}

func NewWordGroupRepository(db *sql.DB) *WordGroupRepository {
	return &WordGroupRepository{db: db}
}

func (r *WordGroupRepository) AddWordToGroup(ctx context.Context, wordID, groupID int) error {
	query := `
		INSERT INTO words_groups (word_id, group_id, created_at)
		VALUES (?, ?, ?)
	`
	_, err := database.ExecContext(ctx, r.db, defaultTimeout, query,
		wordID, groupID, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.ErrDBQuery, "failed to add word to group")
	}
	return nil
}

func (r *WordGroupRepository) RemoveWordFromGroup(ctx context.Context, wordID, groupID int) error {
	query := `DELETE FROM words_groups WHERE word_id = ? AND group_id = ?`
	result, err := database.ExecContext(ctx, r.db, defaultTimeout, query, wordID, groupID)
	if err != nil {
		return errors.Wrap(err, errors.ErrDBQuery, "failed to remove word from group")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.ErrDBQuery, "failed to get affected rows")
	}

	if affected == 0 {
		return errors.New(errors.ErrDBNotFound, "word not found in group")
	}

	return nil
}

func (r *WordGroupRepository) GetGroupWords(ctx context.Context, groupID, offset, limit int) ([]*models.Word, error) {
	query := `
		SELECT w.id, w.japanese, w.romaji, w.english, w.parts, w.created_at, w.updated_at
		FROM words w
		JOIN words_groups wg ON w.id = wg.word_id
		WHERE wg.group_id = ?
		ORDER BY w.id LIMIT ? OFFSET ?
	`

	rows, err := database.QueryContext(ctx, r.db, defaultTimeout, query, groupID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDBQuery, "failed to get group words")
	}
	defer rows.Close()

	var words []*models.Word
	for rows.Next() {
		word := &models.Word{}
		err := rows.Scan(
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

func (r *WordGroupRepository) CountGroupWords(ctx context.Context, groupID int) (int, error) {
	query := `SELECT COUNT(*) FROM words_groups WHERE group_id = ?`
	
	rows, err := database.QueryContext(ctx, r.db, defaultTimeout, query, groupID)
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDBQuery, "failed to count group words")
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, errors.New(errors.ErrDBQuery, "failed to get count")
	}

	var count int
	if err := rows.Scan(&count); err != nil {
		return 0, errors.Wrap(err, errors.ErrDBQuery, "failed to scan count")
	}

	return count, nil
} 