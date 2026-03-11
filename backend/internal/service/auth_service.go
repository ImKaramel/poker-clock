package service

import (
	"errors"
	"sync"

	"backend/internal/domain"
)

type AuthService struct{}

// TODO: key by email --- generate uuid
type InMemoryAuthService struct {
	mu    sync.Mutex
	users map[string]domain.User // key by emailб
}

func NewAuthService() *InMemoryAuthService {
	return &InMemoryAuthService{
		users: make(map[string]domain.User),
	}
}

func (s *InMemoryAuthService) Register(email, password string) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[email]; exists {
		return domain.User{}, errors.New("user already exists")
	}

	user := domain.User{
		ID:       email,
		Email:    email,
		Password: password,
	}

	s.users[email] = user

	return user, nil
}

func (s *InMemoryAuthService) Login(email, password string) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[email]
	if !exists {
		return domain.User{}, errors.New("invalid credentials")
	}

	if user.Password != password {
		return domain.User{}, errors.New("invalid credentials")
	}

	return user, nil
}
