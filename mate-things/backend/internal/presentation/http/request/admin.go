package presentationhttprequest

import (
	"encoding/json"
	"time"
)

type PermissionPostRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type PermissionPatchRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type RolePostRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type RolePatchRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type PayloadSchemaPostRequest struct {
	Name       string          `json:"name"`
	Version    int32           `json:"version"`
	Definition json.RawMessage `json:"definition"`
	ValidFrom  *time.Time      `json:"valid_from"`
	ValidTo    *time.Time      `json:"valid_to"`
}

type PayloadSchemaPatchRequest struct {
	Name       *string          `json:"name"`
	Version    *int32           `json:"version"`
	Definition *json.RawMessage `json:"definition"`
	ValidFrom  *time.Time       `json:"valid_from"`
	ValidTo    *time.Time       `json:"valid_to"`
}

type UserPostRequest struct {
	RoleId   string  `json:"role_id"`
	Name     string  `json:"name"`
	Bio      *string `json:"bio"`
	Username string  `json:"username"`
	Password string  `json:"password"`
}

type UserPatchRequest struct {
	RoleId   *string `json:"role_id"`
	Name     *string `json:"name"`
	Bio      *string `json:"bio"`
	Username *string `json:"username"`
}

type UserPasswordPatchRequest struct {
	Password string `json:"password"`
}
