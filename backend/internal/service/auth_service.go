package service

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/repository"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (domain.User, error) {
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	if existing != nil {
		return domain.User{}, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return *user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	if user == nil {
		return domain.User{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, ErrInvalidCredentials
	}

	return *user, nil
}
