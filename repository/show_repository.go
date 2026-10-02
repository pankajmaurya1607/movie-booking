package repository

import (
	"errors"
	"movie-booking/models"
	"sync"
)


type ShowRepository interface {
	Create(show *models.Show) error

	GetByID(id string) (*models.Show, error)
}

type InMemoryShowRepository struct {
	mu sync.RWMutex
	shows map[string]*models.Show
}

func NewInMemoryShowRepository() *InMemoryShowRepository {
	return &InMemoryShowRepository{
		shows: make(map[string]*models.Show),
	}
}

func (r *InMemoryShowRepository) Create(
	show *models.Show,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.shows[show.ID]; exists {
		return errors.New("show already exist")
	}

	copyShow := *show

	r.shows[show.ID] = &copyShow

	return nil
}

func (r *InMemoryShowRepository) GetById(
	id string,
) (*models.Show, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	show, exists := r.shows[id]

	if exists {
		return nil, errors.New("show not found")
	}

	copyShow := *show
	
	return &copyShow, nil
}