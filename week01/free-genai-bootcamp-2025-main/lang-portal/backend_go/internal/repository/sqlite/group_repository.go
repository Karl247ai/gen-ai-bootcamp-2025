package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/your-org/lang-portal/internal/database"
	"github.com/your-org/lang-portal/internal/errors"
	"github.com/your-org/lang-portal/internal/models"
)

type GroupRepository struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(ctx context.Context, group *models.Group) error {
	query := `
		INSERT INTO groups (name, created_at, updated_at)
		VALUES (?, ?, ?)
	`
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now

	result, err := database.ExecContext(ctx, r.db, defaultTimeout, query,
		group.Name,
		group.CreatedAt,
		group.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(err, errors.ErrDBQuery, "failed to create group")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, errors.ErrDBQuery, "failed to get last insert ID")
	}

	group.ID = int(id)
	return nil
}

func (r *GroupRepository) GetByID(ctx context.Context, id int) (*models.Group, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM groups WHERE id = ?
	`

	rows, err := database.QueryContext(ctx, r.db, defaultTimeout, query, id)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDBQuery, "failed to query group")
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New(errors.ErrDBNotFound, "group not found")
	}

	group := &models.Group{}
	err = rows.Scan(
		&group.ID,
		&group.Name,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDBQuery, "failed to scan group")
	}

	return group, nil
}

func (r *GroupRepository) List(ctx context.Context, offset, limit int) ([]*models.Group, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM groups ORDER BY id LIMIT ? OFFSET ?
	`

	rows, err := database.QueryContext(ctx, r.db, defaultTimeout, query, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDBQuery, "failed to list groups")
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		group := &models.Group{}
		err = rows.Scan(
			&group.ID,
			&group.Name,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDBQuery, "failed to scan group")
		}
		groups = append(groups, group)
	}

	return groups, nil
} 