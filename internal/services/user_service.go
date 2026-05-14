package services

import (
	"context"
	"errors"
	"panzucha/internal/auth"
	"panzucha/internal/domain"
)

type UserService interface {
	Register(ctx context.Context, email, name, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error) // returns JWT token
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, id, email, name string) (*domain.User, error)
}

type userService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, email, name, password string) (*domain.User, error) {
	// Check if user already exists
	existing, _ := s.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	user := &domain.User{
		ID:       domain.NewUserID(),
		Email:    email,
		Name:     name,
		Password: password,
	}

	// Validate and hash password
	if err := user.ValidateForCreate(); err != nil {
		return nil, err
	}
	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Remove password before returning (optional)
	user.Password = ""
	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if user == nil || !user.CheckPassword(password) {
		return "", errors.New("invalid credentials")
	}

	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	// Never send password hash to client
	user.Password = ""
	return user, nil
}

func (s *userService) Update(ctx context.Context, id, email, name string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	user.Email = email
	user.Name = name

	if err := user.ValidateForUpdate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}
