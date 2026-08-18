package services_test

import (
	"context"
	"testing"

	"panzucha/internal/order/domain"
	"panzucha/internal/order/services"
)

type mockUserRepository struct {
	usersByID    map[string]*domain.User
	usersByEmail map[string]*domain.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersByID:    make(map[string]*domain.User),
		usersByEmail: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepository) Create(ctx context.Context, u *domain.User) error {
	m.usersByID[u.ID] = u
	m.usersByEmail[u.Email] = u
	return nil
}

func (m *mockUserRepository) Update(ctx context.Context, u *domain.User) error {
	if _, ok := m.usersByID[u.ID]; !ok {
		return domain.ErrNotFound
	}
	m.usersByID[u.ID] = u
	m.usersByEmail[u.Email] = u
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	u, ok := m.usersByID[id]
	if !ok {
		return domain.ErrNotFound
	}
	delete(m.usersByID, id)
	delete(m.usersByEmail, u.Email)
	return nil
}

func TestRegisterAndLogin(t *testing.T) {
	repo := newMockUserRepository()

	svc := services.NewUserService(repo)

	ctx := context.Background()

	// 1. Test Register
	email := "test@example.com"
	name := "Test User"
	password := "securepassword123"

	user, err := svc.Register(ctx, email, name, password)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	if user.Email != email {
		t.Errorf("expected email %q, got %q", email, user.Email)
	}
	if user.Name != name {
		t.Errorf("expected name %q, got %q", name, user.Name)
	}
	if user.PasswordHash == password {
		t.Error("expected password to be hashed, but it was stored in plaintext")
	}

	// 2. Test Duplicate Registration (Conflict)
	_, err = svc.Register(ctx, email, "Another Name", password)
	if err != domain.ErrConflict {
		t.Errorf("expected ErrConflict, got %v", err)
	}

	// 3. Test Successful Login
	token, err := svc.Login(ctx, email, password)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// 4. Test Login with Wrong Password
	_, err = svc.Login(ctx, email, "wrongpassword")
	if err != domain.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}

	// 5. Test Login with Non-Existent User
	_, err = svc.Login(ctx, "nonexistent@example.com", password)
	if err != domain.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}

	// 6. Test Update
	newName := "Updated User Name"
	newEmail := "updated@example.com"
	newUpdatedBy := "admin"
	updatedUser, err := svc.Update(ctx, user.ID, newEmail, newName, newUpdatedBy)
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	if updatedUser.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updatedUser.Name)
	}
	if updatedUser.Email != newEmail {
		t.Errorf("expected email %q, got %q", newEmail, updatedUser.Email)
	}
}
