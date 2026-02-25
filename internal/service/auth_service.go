package service

import (
	"context"
	"errors"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error)
	Refresh(ctx context.Context, req domain.RefreshRequest) (*domain.AuthResponse, error)
	Logout(ctx context.Context, userID uint) error
	Me(ctx context.Context, userID uint) (*domain.UserResponse, error)
	ListUsers(ctx context.Context) ([]domain.UserResponse, error)
}

type authService struct {
	userRepo     repository.UserRepository
	tokenService TokenService
}

func NewAuthService(userRepo repository.UserRepository, tokenService TokenService) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

func (s *authService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	role := req.Role
	if role == "" {
		role = domain.RolePatient
	}
	if !domain.ValidRole(role) {
		return nil, errors.New("invalid role")
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		MiddleName:   req.MiddleName,
		Phone:        req.Phone,
		Role:         role,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return s.generateTokens(ctx, user)
}

func (s *authService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return s.generateTokens(ctx, user)
}

func (s *authService) Refresh(ctx context.Context, req domain.RefreshRequest) (*domain.AuthResponse, error) {
	userID, err := s.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to find user")
	}

	if user.RefreshToken != req.RefreshToken {
		return nil, errors.New("refresh token revoked")
	}

	return s.generateTokens(ctx, user)
}

func (s *authService) Logout(ctx context.Context, userID uint) error {
	return s.userRepo.UpdateRefreshToken(ctx, userID, "")
}

func (s *authService) Me(ctx context.Context, userID uint) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to find user")
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (s *authService) ListUsers(ctx context.Context) ([]domain.UserResponse, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, errors.New("failed to list users")
	}

	var resp []domain.UserResponse
	for _, u := range users {
		resp = append(resp, u.ToResponse())
	}
	return resp, nil
}

func (s *authService) generateTokens(ctx context.Context, user *domain.User) (*domain.AuthResponse, error) {
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, errors.New("failed to save refresh token")
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}
