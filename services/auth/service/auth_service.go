package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"preppi.com/pkg/auth"
	"preppi.com/services/auth/repository"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrInvalidCreds = errors.New("invalid credentials")
)

type AuthService struct {
	repo  repository.Repository
	token *auth.Manager
}

func New(repo repository.Repository, token *auth.Manager) *AuthService {
	return &AuthService{repo: repo, token: token}
}

func (s *AuthService) Register(ctx context.Context, name, email, password, role, subject string) (uint, string, error) {
	if _, err := s.repo.FindByEmail(ctx, email); err == nil {
		return 0, "", ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", fmt.Errorf("auth_service_hash_password: %w", err)
	}

	u := &repository.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		Subject:      subject,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return 0, "", err
	}

	return u.ID, u.Email, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, uint, string, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", 0, "", ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", 0, "", ErrInvalidCreds
	}

	userID := fmt.Sprintf("%d", u.ID)
	access, err := s.token.GenerateAccessToken(userID, u.Role)
	if err != nil {
		return "", "", 0, "", fmt.Errorf("auth_service_generate_access: %w", err)
	}
	refresh, err := s.token.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", 0, "", fmt.Errorf("auth_service_generate_refresh: %w", err)
	}

	return access, refresh, u.ID, u.Role, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.token.ValidateToken(refreshToken)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	access, err := s.token.GenerateAccessToken(claims.UserID, claims.Role)
	if err != nil {
		return "", fmt.Errorf("auth_service_refresh: %w", err)
	}
	return access, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, userID uint) error {
	return s.repo.MarkEmailVerified(ctx, userID)
}
