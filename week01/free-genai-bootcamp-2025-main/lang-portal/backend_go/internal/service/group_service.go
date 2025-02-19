package service

import (
	"context"
	"database/sql"

	"github.com/your-org/lang-portal/internal/database"
	"github.com/your-org/lang-portal/internal/errors"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/repository"
)

type GroupService interface {
	CreateGroup(ctx context.Context, group *models.Group) error
	GetGroup(ctx context.Context, id int) (*models.Group, error)
	ListGroups(ctx context.Context, page, pageSize int) ([]*models.Group, error)
	AddWordsToGroup(ctx context.Context, groupID int, wordIDs []int) error
}

type groupService struct {
	groupRepo     repository.GroupRepository
	wordGroupRepo repository.WordGroupRepository
	db            *sql.DB
}

func NewGroupService(groupRepo repository.GroupRepository, wordGroupRepo repository.WordGroupRepository, db *sql.DB) GroupService {
	return &groupService{
		groupRepo:     groupRepo,
		wordGroupRepo: wordGroupRepo,
		db:            db,
	}
}

func (s *groupService) CreateGroup(ctx context.Context, group *models.Group) error {
	if group.Name == "" {
		return errors.New(errors.ErrInvalidInput, "group name is required")
	}

	return s.groupRepo.Create(ctx, group)
}

func (s *groupService) GetGroup(ctx context.Context, id int) (*models.Group, error) {
	if id <= 0 {
		return nil, errors.New(errors.ErrInvalidInput, "invalid group ID")
	}

	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) ListGroups(ctx context.Context, page, pageSize int) ([]*models.Group, error) {
	if page < 1 || pageSize < 1 {
		return nil, errors.New(errors.ErrInvalidInput, "invalid pagination parameters")
	}

	offset := (page - 1) * pageSize
	return s.groupRepo.List(ctx, offset, pageSize)
}

// AddWordsToGroup adds multiple words to a group in a single transaction
func (s *groupService) AddWordsToGroup(ctx context.Context, groupID int, wordIDs []int) error {
	if len(wordIDs) == 0 {
		return errors.New(errors.ErrInvalidInput, "no words provided")
	}

	return database.TransactionContext(ctx, s.db, func(tx *sql.Tx) error {
		// Verify group exists
		group, err := s.groupRepo.GetByID(ctx, groupID)
		if err != nil {
			return err
		}
		if group == nil {
			return errors.New(errors.ErrDBNotFound, "group not found")
		}

		// Add each word to the group
		for _, wordID := range wordIDs {
			if err := s.wordGroupRepo.AddWordToGroup(ctx, wordID, groupID); err != nil {
				return errors.Wrap(err, errors.ErrDBTransaction, "failed to add word to group")
			}
		}

		return nil
	})
}

func (s *groupService) UpdateGroup(group *models.Group) error {
	return s.groupRepo.Update(group)
}

func (s *groupService) DeleteGroup(id int) error {
	return s.groupRepo.Delete(id)
} 