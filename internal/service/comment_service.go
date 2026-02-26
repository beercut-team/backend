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
	repo        repository.CommentRepository
	patientRepo repository.PatientRepository
	userRepo    repository.UserRepository
	notifRepo   repository.NotificationRepository
}

func NewCommentService(repo repository.CommentRepository, patientRepo repository.PatientRepository, userRepo repository.UserRepository, notifRepo repository.NotificationRepository) CommentService {
	return &commentService{repo: repo, patientRepo: patientRepo, userRepo: userRepo, notifRepo: notifRepo}
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
		return nil, errors.New("не удалось создать комментарий")
	}

	// Создать уведомление для пациента о новом комментарии
	if s.notifRepo != nil && s.patientRepo != nil && s.userRepo != nil {
		patient, err := s.patientRepo.FindByID(ctx, req.PatientID)
		if err == nil && patient.DoctorID != 0 {
			author, _ := s.userRepo.FindByID(ctx, authorID)
			authorName := "Врач"
			if author != nil {
				authorName = author.Name
			}

			// Уведомление для врача пациента (если комментарий не от него)
			if patient.DoctorID != authorID {
				s.notifRepo.Create(ctx, &domain.Notification{
					UserID:     patient.DoctorID,
					Type:       domain.NotifNewComment,
					Title:      "Новый комментарий",
					Body:       authorName + " добавил комментарий к пациенту " + patient.LastName + " " + patient.FirstName,
					EntityType: "comment",
					EntityID:   comment.ID,
				})
			}

			// Уведомление для хирурга (если назначен и комментарий не от него)
			if patient.SurgeonID != nil && *patient.SurgeonID != authorID {
				s.notifRepo.Create(ctx, &domain.Notification{
					UserID:     *patient.SurgeonID,
					Type:       domain.NotifNewComment,
					Title:      "Новый комментарий",
					Body:       authorName + " добавил комментарий к пациенту " + patient.LastName + " " + patient.FirstName,
					EntityType: "comment",
					EntityID:   comment.ID,
				})
			}
		}
	}

	return comment, nil
}

func (s *commentService) GetByPatient(ctx context.Context, patientID uint) ([]domain.Comment, error) {
	return s.repo.FindByPatient(ctx, patientID)
}

func (s *commentService) MarkAsRead(ctx context.Context, patientID, userID uint) error {
	return s.repo.MarkAsRead(ctx, patientID, userID)
}
