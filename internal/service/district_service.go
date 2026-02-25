package service

import (
	"context"
	"errors"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"gorm.io/gorm"
)

type DistrictService interface {
	Create(ctx context.Context, req domain.CreateDistrictRequest) (*domain.District, error)
	GetByID(ctx context.Context, id uint) (*domain.District, error)
	List(ctx context.Context, search string, offset, limit int) ([]domain.District, int64, error)
	Update(ctx context.Context, id uint, req domain.UpdateDistrictRequest) (*domain.District, error)
	Delete(ctx context.Context, id uint) error
}

type districtService struct {
	repo repository.DistrictRepository
}

func NewDistrictService(repo repository.DistrictRepository) DistrictService {
	return &districtService{repo: repo}
}

func (s *districtService) Create(ctx context.Context, req domain.CreateDistrictRequest) (*domain.District, error) {
	tz := req.Timezone
	if tz == "" {
		tz = "Europe/Moscow"
	}
	district := &domain.District{
		Name:     req.Name,
		Region:   req.Region,
		Code:     req.Code,
		Timezone: tz,
	}
	if err := s.repo.Create(ctx, district); err != nil {
		return nil, errors.New("не удалось создать район")
	}
	return district, nil
}

func (s *districtService) GetByID(ctx context.Context, id uint) (*domain.District, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("район не найден")
		}
		return nil, err
	}
	return d, nil
}

func (s *districtService) List(ctx context.Context, search string, offset, limit int) ([]domain.District, int64, error) {
	return s.repo.FindAll(ctx, search, offset, limit)
}

func (s *districtService) Update(ctx context.Context, id uint, req domain.UpdateDistrictRequest) (*domain.District, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("район не найден")
		}
		return nil, err
	}

	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Region != nil {
		d.Region = *req.Region
	}
	if req.Code != nil {
		d.Code = *req.Code
	}
	if req.Timezone != nil {
		d.Timezone = *req.Timezone
	}

	if err := s.repo.Update(ctx, d); err != nil {
		return nil, errors.New("не удалось обновить район")
	}
	return d, nil
}

func (s *districtService) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("район не найден")
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}
