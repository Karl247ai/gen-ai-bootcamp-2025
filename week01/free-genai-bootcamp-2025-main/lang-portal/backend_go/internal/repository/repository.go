package repository

import (
	"context"
	"github.com/your-org/lang-portal/internal/models"
)

type WordRepository interface {
	Create(ctx context.Context, word *models.Word) error
	GetByID(ctx context.Context, id int) (*models.Word, error)
	List(ctx context.Context, offset, limit int) ([]*models.Word, error)
	Update(word *models.Word) error
	Delete(id int) error
}

type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) error
	GetByID(ctx context.Context, id int) (*models.Group, error)
	List(ctx context.Context, offset, limit int) ([]*models.Group, error)
	Update(group *models.Group) error
	Delete(id int) error
}

type WordGroupRepository interface {
	AddWordToGroup(ctx context.Context, wordID, groupID int) error
	RemoveWordFromGroup(ctx context.Context, wordID, groupID int) error
	GetGroupWords(ctx context.Context, groupID, offset, limit int) ([]*models.Word, error)
	CountGroupWords(ctx context.Context, groupID int) (int, error)
} 