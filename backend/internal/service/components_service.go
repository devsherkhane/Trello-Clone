package service

import (
	"fmt"
	"github.com/devsherkhane/drift/internal/models"
	"github.com/devsherkhane/drift/internal/repository"
	"github.com/devsherkhane/drift/internal/utils"
)

type ComponentService interface {
	GetLists(boardID int) ([]models.List, error)
	GetCards(listID int) ([]models.Card, error)
}

type componentService struct {
	listRepo repository.ListRepository
	cardRepo repository.CardRepository
}

func NewComponentService(listRepo repository.ListRepository, cardRepo repository.CardRepository) ComponentService {
	return &componentService{listRepo: listRepo, cardRepo: cardRepo}
}

func (s *componentService) GetLists(boardID int) ([]models.List, error) {
	return s.listRepo.GetByBoardID(boardID)
}

func (s *componentService) GetCards(listID int) ([]models.Card, error) {
	return s.cardRepo.GetByListID(listID)
}

type ListService interface {
	CreateList(boardID, userID int, title string) (*models.List, error)
	GetListsByBoard(boardID int) ([]models.List, error)
	UpdateListTitle(listID, userID int, title string) error
	DeleteList(listID, userID int) error
}

type listService struct {
	listRepo repository.ListRepository
}

func NewListService(repo repository.ListRepository) ListService {
	return &listService{listRepo: repo}
}

func (s *listService) CreateList(boardID, userID int, title string) (*models.List, error) {
	id, err := s.listRepo.Create(boardID, title)
	if err == nil {
		utils.LogActivity(userID, boardID, "created list: "+title)
	}
	if err != nil {
		return nil, err
	}
	return &models.List{ID: int(id), BoardID: boardID, Title: title}, nil
}

func (s *listService) GetListsByBoard(boardID int) ([]models.List, error) {
	return s.listRepo.GetByBoardID(boardID)
}

func (s *listService) UpdateListTitle(listID, userID int, title string) error {
	err := s.listRepo.UpdateTitle(listID, title)
	if err == nil {
		list, _ := s.listRepo.GetByID(listID)
		if list != nil {
			utils.LogActivity(userID, list.BoardID, "renamed list to "+title)
		}
	}
	return err
}

func (s *listService) DeleteList(listID, userID int) error {
	list, _ := s.listRepo.GetByID(listID)
	err := s.listRepo.Delete(listID)
	if err == nil && list != nil {
		utils.LogActivity(userID, list.BoardID, "deleted list: "+list.Title)
	}
	return err
}

type CardService interface {
	CreateCard(listID, userID int, title string) (*models.Card, error)
	GetCardsByList(listID int) ([]models.Card, error)
	GetCardByID(cardID int) (*models.Card, error)
	UpdateCard(card *models.Card, userID int) error
	MoveCard(cardID, userID, newListID, newPosition int) error
	DeleteCard(cardID, userID int) error
}

type cardService struct {
	cardRepo repository.CardRepository
	listRepo repository.ListRepository
}

func NewCardService(repo repository.CardRepository, listRepo repository.ListRepository) CardService {
	return &cardService{cardRepo: repo, listRepo: listRepo}
}

func (s *cardService) CreateCard(listID, userID int, title string) (*models.Card, error) {
	id, err := s.cardRepo.Create(listID, title)
	if err != nil {
		return nil, err
	}
	card, err := s.cardRepo.GetByID(int(id))
	if err == nil {
		list, _ := s.listRepo.GetByID(listID)
		if list != nil {
			utils.LogActivity(userID, list.BoardID, "created card: "+title)
		}
	}
	return card, err
}

func (s *cardService) GetCardsByList(listID int) ([]models.Card, error) {
	return s.cardRepo.GetByListID(listID)
}

func (s *cardService) GetCardByID(cardID int) (*models.Card, error) {
	return s.cardRepo.GetByID(cardID)
}

func (s *cardService) UpdateCard(card *models.Card, userID int) error {
	err := s.cardRepo.Update(card)
	if err == nil {
		list, _ := s.listRepo.GetByID(card.ListID)
		if list != nil {
			utils.LogActivity(userID, list.BoardID, "updated card: "+card.Title)
		}
	}
	return err
}

func (s *cardService) MoveCard(cardID, userID, newListID, newPosition int) error {
	card, _ := s.cardRepo.GetByID(cardID)
	targetList, _ := s.listRepo.GetByID(newListID)

	err := s.cardRepo.Move(cardID, newListID, newPosition)
	if err == nil && card != nil && targetList != nil {
		utils.LogActivity(userID, targetList.BoardID, fmt.Sprintf("moved card '%s' to list '%s'", card.Title, targetList.Title))
	}
	return err
}

func (s *cardService) DeleteCard(cardID, userID int) error {
	card, _ := s.cardRepo.GetByID(cardID)
	err := s.cardRepo.Delete(cardID)
	if err == nil && card != nil {
		// Need boardID from listID
		list, _ := s.listRepo.GetByID(card.ListID)
		if list != nil {
			utils.LogActivity(userID, list.BoardID, "deleted card: "+card.Title)
		}
	}
	return err
}
