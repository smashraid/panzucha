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

	if err := user.ValidateForCreate(); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityCreate,
			EntityType:  "user",
			EntityID:    user.ID,
			Message:     logger.MsgBusinessValidationFailed,
			Err:         err,
		})
		return nil, err
	}
	if err := user.HashPassword(); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityCreate,
			EntityType:  "user",
			EntityID:    user.ID,
			Message:     "failed to hash password",
			Err:         err,
		})
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityCreate,
			EntityType:  "user",
			EntityID:    user.ID,
			Message:     logger.MsgBusinessCreateFailed,
			Err:         err,
		})
		return nil, err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityCreate,
		EntityType:  "user",
		EntityID:    user.ID,
		Message:     logger.MsgBusinessCreated,
		Err:         nil,
	})
	user.Password = ""
	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "user",
			EntityID:    email,
			Message:     logger.MsgBusinessNotFound,
			Err:         err,
		})
		return "", errors.New("invalid credentials")
	}

	if !user.CheckPassword(password) {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "user",
			EntityID:    user.ID,
			Message:     "invalid password",
			Err:         nil,
		})
		return "", errors.New("invalid credentials")
	}

	roles := []string{user.Role}
	token, err := auth.GenerateToken(user.ID, user.Email, roles)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "user",
			EntityID:    user.ID,
			Message:     "token generation failed",
			Err:         err,
		})
		return "", err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityGet,
		EntityType:  "user",
		EntityID:    user.ID,
		Message:     "user logged in successfully",
		Err:         nil,
	})
	return token, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "user",
			EntityID:    user.ID,
			Message:     logger.MsgBusinessDatabaseError,
			Err:         err,
		})
		return nil, err
	}
	if user == nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "user",
			EntityID:    id,
			Message:     logger.MsgBusinessNotFound,
			Err:         nil,
		})
		return nil, errors.New(logger.MsgBusinessNotFound)
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityGet,
		EntityType:  "user",
		EntityID:    user.ID,
		Message:     logger.MsgBusinessGetFailed,
		Err:         nil,
	})
	user.Password = ""
	return user, nil
}

func (s *userService) Update(ctx context.Context, id, email, name string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.New(logger.MsgBusinessNotFound)
	}

	oldEmail := user.Email
	oldName := user.Name
	user.Email = email
	user.Name = name

	if err := user.ValidateForUpdate(); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "user",
			EntityID:    id,
			Message:     logger.MsgBusinessValidationFailed,
			Err:         err,
		})
		return nil, err
	}

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "user",
			EntityID:    id,
			Message:     logger.MsgBusinessUpdateFailed,
			Err:         err,
		})
		return nil, err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityUpdate,
		EntityType:  "user",
		EntityID:    id,
		Message:     "Data Updated (email: " + oldEmail + "→" + email + ", name: " + oldName + "→" + name + ")",
		Err:         nil,
	})

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityUpdate,
		EntityType:  "user",
		EntityID:    id,
		Message:     logger.MsgBusinessUpdated,
		Err:         nil,
	})
	user.Password = ""
	return user, nil
}
