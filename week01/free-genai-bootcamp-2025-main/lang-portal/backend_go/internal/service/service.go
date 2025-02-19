package service

import "github.com/your-org/lang-portal/internal/models"

type WordService interface {
	CreateWord(word *models.Word) error
	GetWord(id int) (*models.Word, error)
	ListWords(page, pageSize int) ([]*models.Word, error)
	UpdateWord(word *models.Word) error
	DeleteWord(id int) error
}

type GroupService interface {
	CreateGroup(group *models.Group) error
	GetGroup(id int) (*models.Group, error)
	ListGroups(page, pageSize int) ([]*models.Group, error)
	UpdateGroup(group *models.Group) error
	DeleteGroup(id int) error
} 