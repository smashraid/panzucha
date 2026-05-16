package services

import (
	"context"
	"errors"
	"panzucha/internal/auth"
	"panzucha/internal/domain"
	"panzucha/internal/logger"
	"time"
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
	// Check if user already exists
	existing, _ := s.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	user := &domain.User{
		ID:        domain.NewUserID(),
		Email:     email,
		Name:      name,
		Password:  password,
		Role:      "user",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Validate and hash password
	if err := user.ValidateForCreate(); err != nil {
		s.logger.LogBusiness(ctx, "user_registration", "user", user.ID, err.Error(), err)
		return nil, err
	}
	if err := user.HashPassword(); err != nil {
		s.logger.LogBusiness(ctx, "user_registration", "user", user.ID, "failed to hash password", err)
		return nil, err
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.LogBusiness(ctx, "user_registration", "user", user.ID, "database create failed", err)
		return nil, err
	}

	s.logger.LogBusiness(ctx, "user_registration", "user", user.ID, "user registered successfully", nil)
	user.Password = ""
	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		s.logger.LogBusiness(ctx, "user_login", "user", "", "user not found or db error", err)
		return "", errors.New("invalid credentials")
	}

	if !user.CheckPassword(password) {
		s.logger.LogBusiness(ctx, "user_login", "user", user.ID, "invalid password", nil)
		return "", errors.New("invalid credentials")
	}

	roles := []string{user.Role}
	token, err := auth.GenerateToken(user.ID, user.Email, roles)
	if err != nil {
		s.logger.LogBusiness(ctx, "user_login", "user", user.ID, "token generation failed", err)
		return "", err
	}

	s.logger.LogBusiness(ctx, "user_login", "user", user.ID, "user logged in successfully", nil)
	return token, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.LogBusiness(ctx, "user_get", "user", id, "database error", err)
		return nil, err
	}
	if user == nil {
		s.logger.LogBusiness(ctx, "user_get", "user", id, "user not found", nil)
		return nil, errors.New("user not found")
	}
	s.logger.LogBusiness(ctx, "user_get", "user", id, "user retrieved", nil)
	user.Password = ""
	return user, nil
}

func (s *userService) Update(ctx context.Context, id, email, name string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	oldEmail := user.Email
	oldName := user.Name
	user.Email = email
	user.Name = name

	if err := user.ValidateForUpdate(); err != nil {
		s.logger.LogBusiness(ctx, "user_update", "user", id, "user not found or db error", err)
		return nil, err
	}

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.LogBusiness(ctx, "user_update", "user", id, "database update failed", err)
		return nil, err
	}

	s.logger.LogBusiness(ctx, "user_update", "user", id,
		"user updated (email: "+oldEmail+"→"+email+", name: "+oldName+"→"+name+")", nil)
	user.Password = ""
	return user, nil
}
