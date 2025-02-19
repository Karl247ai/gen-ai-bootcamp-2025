package service

import (
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/repository"
)

type WordGroupService interface {
	AddWordToGroup(wordID, groupID int) error
	RemoveWordFromGroup(wordID, groupID int) error
	GetGroupWords(groupID, page, pageSize int) ([]*models.Word, error)
	CountGroupWords(groupID int) (int, error)
}

type wordGroupService struct {
	wordGroupRepo repository.WordGroupRepository
	wordRepo      repository.WordRepository
	groupRepo     repository.GroupRepository
}

func NewWordGroupService(
	wordGroupRepo repository.WordGroupRepository,
	wordRepo repository.WordRepository,
	groupRepo repository.GroupRepository,
) WordGroupService {
	return &wordGroupService{
		wordGroupRepo: wordGroupRepo,
		wordRepo:      wordRepo,
		groupRepo:     groupRepo,
	}
}

func (s *wordGroupService) AddWordToGroup(wordID, groupID int) error {
	// Verify word exists
	word, err := s.wordRepo.GetByID(wordID)
	if err != nil {
		return err
	}
	if word == nil {
		return ErrWordNotFound
	}

	// Verify group exists
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}

	return s.wordGroupRepo.AddWordToGroup(wordID, groupID)
}

func (s *wordGroupService) RemoveWordFromGroup(wordID, groupID int) error {
	return s.wordGroupRepo.RemoveWordFromGroup(wordID, groupID)
}

func (s *wordGroupService) GetGroupWords(groupID, page, pageSize int) ([]*models.Word, error) {
	offset := (page - 1) * pageSize
	return s.wordGroupRepo.GetGroupWords(groupID, offset, pageSize)
}

func (s *wordGroupService) CountGroupWords(groupID int) (int, error) {
	return s.wordGroupRepo.CountGroupWords(groupID)
} 