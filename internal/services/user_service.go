package services

import (
	"context"
	"panzucha/internal/auth"
	"panzucha/internal/domain"
	"panzucha/internal/logger"
)

type UserService interface {
	Register(ctx context.Context, email, name, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error) // returns JWT token
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, id, email, name string) (*domain.User, error)
}

type userService struct {
	repo   domain.UserRepository
	logger *logger.Logger
}

func NewUserService(repo domain.UserRepository, log *logger.Logger) UserService {
	return &userService{repo: repo, logger: log}
}

func (s *userService) Register(ctx context.Context, email, name, password string) (*domain.User, error) {
	existing, _ := s.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, domain.ErrConflict
	}
	user := &domain.User{
		ID:           domain.NewUserID(),
		Email:        email,
		Name:         name,
		PasswordHash: password,
		Role:         "user",
		Audit:        domain.Audit{},
	}

	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return "", domain.ErrUnauthorized
	}

	if !user.CheckPassword(password) {
		return "", domain.ErrUnauthorized
	}

	roles := []string{user.Role}
	token, err := auth.GenerateToken(user.ID, user.Email, roles)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) Update(ctx context.Context, id, email, name string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, err
	}

	user.Email = email
	user.Name = name

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
