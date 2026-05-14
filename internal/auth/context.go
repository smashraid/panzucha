package auth

import "context"

type contextKey string

const (
	UserIDKey contextKey = "userID"
	RolesKey  contextKey = "roles"
)

func ContextWithUser(ctx context.Context, userID string, roles []string) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, RolesKey, roles)
	return ctx
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func RolesFromContext(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(RolesKey).([]string)
	return roles, ok
}
