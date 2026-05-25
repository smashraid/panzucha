package services

import (
	"context"
	"panzucha/internal/domain"
	"time"
)

type IdempotencyService interface {
	// Reserve attempts to insert a processing key. Returns:
	// - isOwner: true if this request inserted the key (first to claim)
	// - cached: the existing completed key (if any) – nil if not found or still processing
	// - err: database error
	Reserve(ctx context.Context, key string) (isOwner bool, cached *domain.IdempotencyKey, err error)

	// Complete updates a processing key to completed with the final response.
	Complete(ctx context.Context, key, resourceType, resourceID string, statusCode int, responseBody []byte) error
	Delete(ctx context.Context, key string) error
}

type idempotencyService struct {
	repo domain.IdempotencyKeyRepository
}

func NewIdempotencyService(repo domain.IdempotencyKeyRepository) IdempotencyService {
	return &idempotencyService{repo: repo}
}

func (s *idempotencyService) Reserve(ctx context.Context, keyStr string) (bool, *domain.IdempotencyKey, error) {
	// Try to insert a processing row
	now := time.Now().UTC()
	key := &domain.IdempotencyKey{
		Key:       keyStr,
		Status:    domain.IdempotencyStatusProcessing,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	err := s.repo.Create(ctx, key)
	if err != nil {
		// Check if duplicate key (already exists)
		// pgx unique violation error code is "23505"
		if isUniqueViolation(err) {
			// Key already exists – fetch it
			existing, fetchErr := s.repo.FindByKey(ctx, keyStr)
			if fetchErr != nil {
				return false, nil, fetchErr
			}
			if existing == nil {
				// Should not happen, but treat as not owned
				return false, nil, nil
			}
			if existing.Status == domain.IdempotencyStatusCompleted {
				return false, existing, nil
			}
			// Still processing – not owned by this request
			return false, nil, nil
		}
		return false, nil, err
	}
	// Successfully inserted – this request owns the key
	return true, nil, nil
}

func (s *idempotencyService) Complete(ctx context.Context, key, resourceType, resourceID string, statusCode int, responseBody []byte) error {
	return s.repo.UpdateToCompleted(ctx, key, resourceID, statusCode, responseBody)
}

func (s *idempotencyService) Delete(ctx context.Context, key string) error {
	return s.repo.Delete(ctx, key)
}

func isUniqueViolation(err error) bool {
	// pgx error code for unique violation
	if pqErr, ok := err.(interface{ Code() string }); ok {
		return pqErr.Code() == "23505"
	}
	return false
}
