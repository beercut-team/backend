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
	PatientLogin(ctx context.Context, req domain.PatientLoginRequest) (*domain.AuthResponse, error)
	Refresh(ctx context.Context, req domain.RefreshRequest) (*domain.AuthResponse, error)
	Logout(ctx context.Context, userID uint) error
	Me(ctx context.Context, userID uint) (*domain.UserResponse, error)
	ListUsers(ctx context.Context) ([]domain.UserResponse, error)
}

type authService struct {
	userRepo     repository.UserRepository
	patientRepo  repository.PatientRepository
	tokenService TokenService
}

func NewAuthService(userRepo repository.UserRepository, tokenService TokenService) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

func NewAuthServiceWithPatient(userRepo repository.UserRepository, patientRepo repository.PatientRepository, tokenService TokenService) AuthService {
	return &authService{
		userRepo:     userRepo,
		patientRepo:  patientRepo,
		tokenService: tokenService,
	}
}

func (s *authService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("этот email уже зарегистрирован")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("ошибка хеширования пароля")
	}

	role := req.Role
	if role == "" {
		role = domain.RolePatient
	}
	if !domain.ValidRole(role) {
		return nil, errors.New("недопустимая роль")
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
		return nil, errors.New("не удалось создать пользователя")
	}

	return s.generateTokens(ctx, user)
}

func (s *authService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("неверный email или пароль")
	}

	if !user.IsActive {
		return nil, errors.New("аккаунт деактивирован")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("неверный email или пароль")
	}

	return s.generateTokens(ctx, user)
}

func (s *authService) PatientLogin(ctx context.Context, req domain.PatientLoginRequest) (*domain.AuthResponse, error) {
	if s.patientRepo == nil {
		return nil, errors.New("вход по коду доступа недоступен")
	}

	// Find patient by access code (case-insensitive)
	patient, err := s.patientRepo.FindByAccessCode(ctx, req.AccessCode)
	if err != nil {
		return nil, errors.New("неверный код доступа")
	}

	// Generate tokens with patient ID and PATIENT role
	accessToken, err := s.tokenService.GenerateAccessToken(patient.ID, domain.RolePatient)
	if err != nil {
		return nil, errors.New("не удалось сгенерировать токен доступа")
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(patient.ID)
	if err != nil {
		return nil, errors.New("не удалось сгенерировать токен обновления")
	}

	// Create virtual user response from patient data
	userResp := domain.UserResponse{
		ID:         patient.ID,
		Email:      patient.Email,
		Name:       patient.FirstName + " " + patient.LastName,
		FirstName:  patient.FirstName,
		LastName:   patient.LastName,
		MiddleName: patient.MiddleName,
		Phone:      patient.Phone,
		Role:       domain.RolePatient,
		IsActive:   true,
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResp,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, req domain.RefreshRequest) (*domain.AuthResponse, error) {
	userID, err := s.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("недействительный токен обновления")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, errors.New("не удалось найти пользователя")
	}

	if user.RefreshToken != req.RefreshToken {
		return nil, errors.New("токен обновления отозван")
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
			return nil, errors.New("пользователь не найден")
		}
		return nil, errors.New("не удалось найти пользователя")
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (s *authService) ListUsers(ctx context.Context) ([]domain.UserResponse, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, errors.New("не удалось получить список пользователей")
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
		return nil, errors.New("не удалось сгенерировать токен доступа")
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("не удалось сгенерировать токен обновления")
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, errors.New("не удалось сохранить токен обновления")
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}
