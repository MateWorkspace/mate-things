package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type RoleResponse struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	IsDefault   bool            `json:"is_default"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

type RolePermissionResponse struct {
	Id           string    `json:"id"`
	RoleId       string    `json:"role_id"`
	PermissionId string    `json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    *string   `json:"created_by,omitempty"`
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
