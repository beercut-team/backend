package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/beercut-team/backend-boilerplate/internal/config"
)

type TokenService interface {
	GenerateAccessToken(userID uint) (string, error)
	GenerateRefreshToken(userID uint) (string, error)
	ValidateAccessToken(tokenStr string) (uint, error)
	ValidateRefreshToken(tokenStr string) (uint, error)
}

type tokenService struct {
	cfg *config.Config
}

func NewTokenService(cfg *config.Config) TokenService {
	return &tokenService{cfg: cfg}
}

func (s *tokenService) GenerateAccessToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
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

func (s *tokenService) ValidateAccessToken(tokenStr string) (uint, error) {
	return s.validateToken(tokenStr, s.cfg.JWTAccessSecret, "access")
}

func (s *tokenService) ValidateRefreshToken(tokenStr string) (uint, error) {
	return s.validateToken(tokenStr, s.cfg.JWTRefreshSecret, "refresh")
}

func (s *tokenService) validateToken(tokenStr, secret, expectedType string) (uint, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token claims")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != expectedType {
		return 0, fmt.Errorf("invalid token type")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid user_id in token")
	}

	return uint(userIDFloat), nil
}
