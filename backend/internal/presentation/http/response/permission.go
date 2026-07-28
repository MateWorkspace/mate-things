package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type PermissionResponse struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

func Permission(permission domainmodels.Permission) PermissionResponse {
	return PermissionResponse{
		Id:          UUIDString(permission.Id),
		Name:        permission.Name,
		Description: permission.Description,
		Preferences: NormalizeJSON(permission.Preferences),
		AuditResponse: Audit(
			permission.CreatedAt,
			permission.UpdatedAt,
			permission.DeletedAt,
			permission.CreatedBy,
			permission.UpdatedBy,
			permission.DeletedBy,
		),
	}
}

func Permissions(permissions []domainmodels.Permission) []PermissionResponse {
	result := make([]PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		result = append(result, Permission(permission))
	}
	return result
}
