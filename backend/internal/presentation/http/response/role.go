package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type RoleResponse struct {
	Id          string          `json:"id" example:"7a3c2e10-4b1a-4b8e-9f3a-2b6e7c9d1a04"`
	Name        string          `json:"name" example:"barista"`
	Description string          `json:"description" example:"Can dispatch brew actions on the shop floor, but can't manage the fleet."`
	IsDefault   bool            `json:"is_default" example:"false"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

type RolePermissionResponse struct {
	Id           string    `json:"id" example:"c8a1d4e7-3f6b-4c9e-8a2d-5b7e0f3c6a09"`
	RoleId       string    `json:"role_id" example:"7a3c2e10-4b1a-4b8e-9f3a-2b6e7c9d1a04"`
	PermissionId string    `json:"permission_id" example:"1d2e3f4a-5b6c-4d7e-8f9a-0b1c2d3e4f5a"`
	CreatedAt    time.Time `json:"created_at" example:"2026-06-15T09:30:00Z"`
	CreatedBy    *string   `json:"created_by,omitempty" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
}

type RolePermissionDetailResponse struct {
	RolePermission RolePermissionResponse `json:"role_permission"`
	Role           RoleResponse           `json:"role"`
	Permission     PermissionResponse     `json:"permission"`
}

func Role(role domainmodels.Role) RoleResponse {
	return RoleResponse{
		Id:          UUIDString(role.Id),
		Name:        role.Name,
		Description: role.Description,
		IsDefault:   role.IsDefault,
		Preferences: NormalizeJSON(role.Preferences),
		AuditResponse: Audit(
			role.CreatedAt,
			role.UpdatedAt,
			role.DeletedAt,
			role.CreatedBy,
			role.UpdatedBy,
			role.DeletedBy,
		),
	}
}

func Roles(roles []domainmodels.Role) []RoleResponse {
	result := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		result = append(result, Role(role))
	}
	return result
}

func RolePermission(rolePermission domainmodels.RolePermission) RolePermissionResponse {
	return RolePermissionResponse{
		Id:           UUIDString(rolePermission.Id),
		RoleId:       UUIDString(rolePermission.RoleId),
		PermissionId: UUIDString(rolePermission.PermissionId),
		CreatedAt:    rolePermission.CreatedAt,
		CreatedBy:    UUIDPtrString(rolePermission.CreatedBy),
	}
}

func RolePermissionDetail(rolePermission domainmodels.RolePermission, role domainmodels.Role, permission domainmodels.Permission) RolePermissionDetailResponse {
	return RolePermissionDetailResponse{
		RolePermission: RolePermission(rolePermission),
		Role:           Role(role),
		Permission:     Permission(permission),
	}
}

func RolePermissionDetails(
	rolePermissions []domainmodels.RolePermission,
	roles []domainmodels.Role,
	permissions []domainmodels.Permission,
) []RolePermissionDetailResponse {
	result := make([]RolePermissionDetailResponse, 0, len(rolePermissions))
	for i, rolePermission := range rolePermissions {
		var role domainmodels.Role
		if i < len(roles) {
			role = roles[i]
		}

		var permission domainmodels.Permission
		if i < len(permissions) {
			permission = permissions[i]
		}

		result = append(result, RolePermissionDetail(rolePermission, role, permission))
	}
	return result
}
