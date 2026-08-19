package handlers

import "panzucha/internal/user/domain"

// User
func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.Audit.CreatedAt,
		UpdatedAt: u.Audit.UpdatedAt,
	}
}
