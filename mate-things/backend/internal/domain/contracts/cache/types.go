package domaincontractscache

import domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"

type Pagination[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

type RolePermissionItem struct {
	RolePermission domainmodels.RolePermission `json:"role_permission"`
	Role           domainmodels.Role           `json:"role"`
	Permission     domainmodels.Permission     `json:"permission"`
}
