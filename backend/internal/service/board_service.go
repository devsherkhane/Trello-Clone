package service

import (
	"github.com/devsherkhane/drift/internal/models"
	"github.com/devsherkhane/drift/internal/repository"
	"github.com/devsherkhane/drift/internal/utils"
)

type BoardService interface {
	CreateBoard(title string, userID int) (*models.Board, error)
	GetBoards(userID int) ([]models.Board, error)
	GetBoardByID(boardID, userID int) (*models.Board, error)
	UpdateBoardTitle(boardID, userID int, title string) error
	DeleteBoard(boardID, userID int) error
	ArchiveBoard(boardID, userID int) error
	GetActivityLogs(boardID int) ([]models.ActivityLog, error)

	AddCollaborator(boardID, userID int, email, role string) error
	GetCollaborators(boardID int) ([]models.User, error)
	UpdateCollaboratorRole(boardID, userID, collaboratorID int, role string) error
	RemoveCollaborator(boardID, userID, collaboratorID int) error
}

type boardService struct {
	boardRepo repository.BoardRepository
}

func NewBoardService(boardRepo repository.BoardRepository) BoardService {
	return &boardService{boardRepo: boardRepo}
}

func (s *boardService) CreateBoard(title string, userID int) (*models.Board, error) {
	id, err := s.boardRepo.Create(title, userID)
	if err != nil {
		return nil, err
	}
	utils.LogActivity(userID, int(id), "created board "+title)
	return &models.Board{ID: int(id), Title: title, UserID: userID}, nil
}

func (s *boardService) GetBoards(userID int) ([]models.Board, error) {
	return s.boardRepo.GetByUserID(userID)
}

func (s *boardService) GetBoardByID(boardID, userID int) (*models.Board, error) {
	return s.boardRepo.GetByID(boardID, userID)
}

func (s *boardService) UpdateBoardTitle(boardID, userID int, title string) error {
	err := s.boardRepo.UpdateTitle(boardID, userID, title)
	if err != nil {
		return err
	}
	utils.LogActivity(userID, boardID, "renamed board to "+title)
	return nil
}

func (s *boardService) DeleteBoard(boardID, userID int) error {
	err := s.boardRepo.Delete(boardID, userID)
	if err == nil {
		utils.LogActivity(userID, boardID, "deleted the board")
	}
	return err
}

func (s *boardService) ArchiveBoard(boardID, userID int) error {
	err := s.boardRepo.Archive(boardID, userID)
	if err == nil {
		utils.LogActivity(userID, boardID, "archived the board")
	}
	return err
}

func (s *boardService) GetActivityLogs(boardID int) ([]models.ActivityLog, error) {
	return s.boardRepo.GetActivityLogs(boardID)
}

func (s *boardService) AddCollaborator(boardID, userID int, email, role string) error {
	// Simple permission check: must be owner to add
	board, err := s.boardRepo.GetByID(boardID, userID)
	if err != nil {
		return err
	}
	if board.UserID != userID {
		return repository.ErrNotFound // Pretend board doesn't exist for unauthorized access
	}

	err = s.boardRepo.AddCollaborator(boardID, email, role)
	if err != nil {
		return err
	}
	utils.LogActivity(userID, boardID, "added "+email+" as "+role)
	return nil
}

func (s *boardService) GetCollaborators(boardID int) ([]models.User, error) {
	return s.boardRepo.GetCollaborators(boardID)
}

func (s *boardService) UpdateCollaboratorRole(boardID, userID, collaboratorID int, role string) error {
	// Verify ownership
	board, err := s.boardRepo.GetByID(boardID, userID)
	if err != nil {
		return err
	}
	if board.UserID != userID {
		return repository.ErrNotFound
	}
	return s.boardRepo.UpdateCollaboratorRole(boardID, collaboratorID, role)
}

func (s *boardService) RemoveCollaborator(boardID, userID, collaboratorID int) error {
	// Verify ownership
	board, err := s.boardRepo.GetByID(boardID, userID)
	if err != nil {
		return err
	}
	if board.UserID != userID {
		return repository.ErrNotFound
	}
	return s.boardRepo.RemoveCollaborator(boardID, collaboratorID)
}
