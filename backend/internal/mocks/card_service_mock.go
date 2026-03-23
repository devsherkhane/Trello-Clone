package mocks

import (
	"github.com/devsherkhane/drift/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockCardService struct {
	mock.Mock
}

func (m *MockCardService) CreateCard(listID, userID int, title string) (*models.Card, error) {
	args := m.Called(listID, userID, title)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Card), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCardService) GetCardsByList(listID int) ([]models.Card, error) {
	args := m.Called(listID)
	if args.Get(0) != nil {
		return args.Get(0).([]models.Card), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCardService) GetCardByID(cardID int) (*models.Card, error) {
	args := m.Called(cardID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Card), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCardService) UpdateCard(card *models.Card, userID int) error {
	args := m.Called(card, userID)
	return args.Error(0)
}

func (m *MockCardService) MoveCard(cardID, userID, newListID, newPosition int) error {
	args := m.Called(cardID, userID, newListID, newPosition)
	return args.Error(0)
}

func (m *MockCardService) DeleteCard(cardID int) error {
	args := m.Called(cardID)
	return args.Error(0)
}
