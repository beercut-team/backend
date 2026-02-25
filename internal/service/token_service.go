package service

import (
	"fmt"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type TokenService interface {
	GenerateAccessToken(userID uint, role domain.Role) (string, error)
	GenerateRefreshToken(userID uint) (string, error)
	ValidateAccessToken(tokenStr string) (uint, domain.Role, error)
	ValidateRefreshToken(tokenStr string) (uint, error)
}

type tokenService struct {
	cfg *config.Config
}

func NewTokenService(cfg *config.Config) TokenService {
	return &tokenService{cfg: cfg}
}

func (s *tokenService) GenerateAccessToken(userID uint, role domain.Role) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    string(role),
		"exp":     time.Now().Add(time.Duration(s.cfg.JWTAccessExpiryMin) * time.Minute).Unix(),
		"type":    "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTAccessSecret))
}

func (s *tokenService) GenerateRefreshToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(s.cfg.JWTRefreshExpiryHrs) * time.Hour).Unix(),
		"type":    "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTRefreshSecret))
}

func (s *tokenService) ValidateAccessToken(tokenStr string) (uint, domain.Role, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTAccessSecret), nil
	})
	if err != nil {
		return 0, "", fmt.Errorf("недействительный токен: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, "", fmt.Errorf("недействительные данные токена")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "access" {
		return 0, "", fmt.Errorf("недействительный тип токена")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, "", fmt.Errorf("недействительный user_id в токене")
	}

	roleStr, _ := claims["role"].(string)

	return uint(userIDFloat), domain.Role(roleStr), nil
}

func (s *tokenService) ValidateRefreshToken(tokenStr string) (uint, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTRefreshSecret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("недействительный токен: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("недействительные данные токена")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return 0, fmt.Errorf("недействительный тип токена")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("недействительный user_id в токене")
	}

	return uint(userIDFloat), nil
}
