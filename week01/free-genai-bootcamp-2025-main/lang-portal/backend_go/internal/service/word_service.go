package service

import (
	"context"

	"github.com/your-org/lang-portal/internal/errors"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/repository"
)

type WordService interface {
	CreateWord(ctx context.Context, word *models.Word) error
	GetWord(ctx context.Context, id int) (*models.Word, error)
	ListWords(ctx context.Context, page, pageSize int) ([]*models.Word, error)
}

type wordService struct {
	wordRepo repository.WordRepository
}

func NewWordService(wordRepo repository.WordRepository) WordService {
	return &wordService{wordRepo: wordRepo}
}

func (s *wordService) CreateWord(ctx context.Context, word *models.Word) error {
	// Validate input
	if word.Japanese == "" || word.English == "" {
		return errors.New(errors.ErrInvalidInput, "japanese and english text are required")
	}

	return s.wordRepo.Create(ctx, word)
}

func (s *wordService) GetWord(ctx context.Context, id int) (*models.Word, error) {
	if id <= 0 {
		return nil, errors.New(errors.ErrInvalidInput, "invalid word ID")
	}

	word, err := s.wordRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return word, nil
}

func (s *wordService) ListWords(ctx context.Context, page, pageSize int) ([]*models.Word, error) {
	if page < 1 || pageSize < 1 {
		return nil, errors.New(errors.ErrInvalidInput, "invalid pagination parameters")
	}

	offset := (page - 1) * pageSize
	return s.wordRepo.List(ctx, offset, pageSize)
}

func (s *wordService) UpdateWord(word *models.Word) error {
	return s.wordRepo.Update(word)
}

func (s *wordService) DeleteWord(id int) error {
	return s.wordRepo.Delete(id)
} 