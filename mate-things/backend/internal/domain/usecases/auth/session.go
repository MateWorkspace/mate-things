package domainusecasesauth

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type Session interface {
	Login(ctx context.Context, request LoginRequest) (LoginResult, error)
	Refresh(ctx context.Context, request RefreshRequest) (LoginResult, error)
}

type LoginRequest struct {
	Username string
	Password string
}

type RefreshRequest struct {
	RefreshToken string
}

type LoginResult struct {
	User         domainmodels.User
	Role         domainmodels.Role
	Permissions  []domainmodels.Permission
	AccessToken  string
	RefreshToken string
}
