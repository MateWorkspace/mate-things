package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type UserResponse struct {
	Id          string          `json:"id"`
	RoleId      string          `json:"role_id"`
	Name        string          `json:"name"`
	Bio         string          `json:"bio"`
	Username    string          `json:"username"`
	Preferences json.RawMessage `json:"preferences"`
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
