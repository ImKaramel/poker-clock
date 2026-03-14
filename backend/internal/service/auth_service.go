package service

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/repository"
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
		return domain.User{}, err
	}
	if existing != nil {
		return domain.User{}, ErrUserExists
	}

	user := &domain.User{
		Email:    email,
		Password: password,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return domain.User{}, err
	}

	return *user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (domain.User, error) {

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	if user == nil {
		return domain.User{}, ErrInvalidCredentials
	}

	if user.Password != password {
		return domain.User{}, ErrInvalidCredentials
	}

	return *user, nil
}
