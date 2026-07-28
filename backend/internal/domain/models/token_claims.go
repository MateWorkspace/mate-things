package domainmodels

import "github.com/google/uuid"

type TokenClaimsAccess struct {
	UserId      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
}

type TokenClaimsRefresh struct {
	UserId uuid.UUID `json:"user_id"`
}
