package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type PermissionResponse struct {
	Id          string          `json:"id" example:"1d2e3f4a-5b6c-4d7e-8f9a-0b1c2d3e4f5a"`
	Name        string          `json:"name" example:"node:get"`
	Description string          `json:"description" example:"View registered nodes and their status."`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
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
