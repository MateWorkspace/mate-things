package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type UserResponse struct {
	Id          string          `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	RoleId      string          `json:"role_id" example:"7a3c2e10-4b1a-4b8e-9f3a-2b6e7c9d1a04"`
	Name        string          `json:"name" example:"Grace Hopper"`
	Bio         string          `json:"bio" example:"Keeps the espresso machines humming and the firmware fresh."`
	Username    string          `json:"username" example:"grace.hopper"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func User(user domainmodels.User) UserResponse {
	return UserResponse{
		Id:          UUIDString(user.Id),
		RoleId:      UUIDString(user.RoleId),
		Name:        user.Name,
		Bio:         user.Bio,
		Username:    user.Username,
		Preferences: NormalizeJSON(user.Preferences),
		AuditResponse: Audit(
			user.CreatedAt,
			user.UpdatedAt,
			user.DeletedAt,
			user.CreatedBy,
			user.UpdatedBy,
			user.DeletedBy,
		),
	}
}

func Users(users []domainmodels.User) []UserResponse {
	result := make([]UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, User(user))
	}
	return result
}
