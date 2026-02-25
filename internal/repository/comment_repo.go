package repository

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	FindByPatient(ctx context.Context, patientID uint) ([]domain.Comment, error)
	FindByID(ctx context.Context, id uint) (*domain.Comment, error)
	MarkAsRead(ctx context.Context, patientID, userID uint) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) FindByPatient(ctx context.Context, patientID uint) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.WithContext(ctx).Where("patient_id = ?", patientID).
		Preload("Author").Order("created_at ASC").Find(&comments).Error
	return comments, err
}

func (r *commentRepository) FindByID(ctx context.Context, id uint) (*domain.Comment, error) {
	var comment domain.Comment
	if err := r.db.WithContext(ctx).Preload("Author").First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) MarkAsRead(ctx context.Context, patientID, userID uint) error {
	return r.db.WithContext(ctx).Model(&domain.Comment{}).
		Where("patient_id = ? AND author_id != ? AND is_read = false", patientID, userID).
		Update("is_read", true).Error
}
