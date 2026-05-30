package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"panzucha/internal/auth"
	"panzucha/internal/domain"
)

type UserService interface {
	Register(ctx context.Context, email, name, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, id, email, name, updatedBy string) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}

type userService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) UserService {
	return &userService{repo: repo}
}

// Register creates a new user account.
//
// Uniqueness check order:
//  1. Service calls GetByEmail — catches the common case with one round-trip.
//  2. If two concurrent requests both pass step 1 before either inserts,
//     the DB UNIQUE constraint fires and the repo translates it to ErrConflict.
//     Both paths surface the same ErrConflict to the handler.
func (s *userService) Register(ctx context.Context, email, name, password string) (*domain.User, error) {
	// Step 1 — check uniqueness at the service layer.
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		// Real DB error — propagate, don't swallow.
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrConflict
	}

	// Step 2 — hash the password in the service, never in the domain entity.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		Name:         name,
		PasswordHash: string(hash),
		Role:         "user",
		Audit:        domain.Audit{CreatedBy: email, UpdatedBy: email},
	}

	if err := s.repo.Create(ctx, user); err != nil {
		// Step 3 — if a concurrent insert beat us, repo returns ErrConflict.
		return nil, err
	}
	return user, nil
}

// Login authenticates a user and returns a signed JWT token.
// Both "user not found" and "wrong password" surface as ErrUnauthorized —
// never reveal which one failed to the client.
func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// Distinguish a real DB error from "user not found".
		// ErrNotFound → wrong credentials → ErrUnauthorized.
		// Any other error → infrastructure failure → propagate.
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrUnauthorized
		}
		return "", err
	}

	// bcrypt comparison lives here, not on the domain entity.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrUnauthorized
	}

	token, err := auth.GenerateToken(user.ID, user.Email, []string{user.Role})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// Update applies email and name changes to an existing user.
// No GetByID before Update — the repository checks existence implicitly
// via RowsAffected and returns ErrNotFound if the user no longer exists.
// This keeps the operation at one DB round-trip on the happy path.
func (s *userService) Update(ctx context.Context, id, email, name, updatedBy string) (*domain.User, error) {
	user := &domain.User{
		ID:    id,
		Email: email,
		Name:  name,
		Audit: domain.Audit{UpdatedBy: updatedBy},
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
