package service

import (
	"context"
	"errors"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
)

type CommentService interface {
	Create(ctx context.Context, req domain.CreateCommentRequest, authorID uint) (*domain.Comment, error)
	GetByPatient(ctx context.Context, patientID uint) ([]domain.Comment, error)
	MarkAsRead(ctx context.Context, patientID, userID uint) error
}

type commentService struct {
	repo repository.CommentRepository
}

func NewCommentService(repo repository.CommentRepository) CommentService {
	return &commentService{repo: repo}
}

func (s *commentService) Create(ctx context.Context, req domain.CreateCommentRequest, authorID uint) (*domain.Comment, error) {
	comment := &domain.Comment{
		PatientID: req.PatientID,
		AuthorID:  authorID,
		ParentID:  req.ParentID,
		Body:      req.Body,
		IsUrgent:  req.IsUrgent,
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, errors.New("failed to create comment")
	}

	return comment, nil
}

func (s *commentService) GetByPatient(ctx context.Context, patientID uint) ([]domain.Comment, error) {
	return s.repo.FindByPatient(ctx, patientID)
}

func (s *commentService) MarkAsRead(ctx context.Context, patientID, userID uint) error {
	return s.repo.MarkAsRead(ctx, patientID, userID)
}
