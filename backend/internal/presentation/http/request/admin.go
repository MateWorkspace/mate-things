package presentationhttprequest

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type PermissionPostRequest struct {
	Name        string  `json:"name" example:"node:get"`
	Description *string `json:"description" example:"View registered nodes and their status."`
}

type PermissionPatchRequest struct {
	Name        *string `json:"name" example:"node:get"`
	Description *string `json:"description" example:"View registered nodes and their status."`
}

type LlmConfigPutRequest struct {
	Provider domainmodels.LlmProvider `json:"provider" example:"CLAUDE"`
	Model    string                   `json:"model" example:"claude-opus-5"`
	ApiKey   string                   `json:"api_key" example:"sk-ant-..."`
	BaseURL  *string                  `json:"base_url,omitempty" example:"https://api.anthropic.com"`
}

type RolePostRequest struct {
	Name        string  `json:"name" example:"barista"`
	Description *string `json:"description" example:"Can dispatch brew actions on the shop floor, but can't manage the fleet."`
}

type RolePatchRequest struct {
	Name        *string `json:"name" example:"barista"`
	Description *string `json:"description" example:"Can dispatch brew actions on the shop floor, but can't manage the fleet."`
}

type PayloadSchemaPostRequest struct {
	Name       string          `json:"name" example:"brew_command"`
	Version    int32           `json:"version" example:"1"`
	Definition json.RawMessage `json:"definition" swaggertype:"object"`
	ValidFrom  *time.Time      `json:"valid_from" example:"2026-07-01T00:00:00Z"`
	ValidTo    *time.Time      `json:"valid_to" example:"2027-07-01T00:00:00Z"`
}

type PayloadSchemaPatchRequest struct {
	Name       *string          `json:"name" example:"brew_command"`
	Version    *int32           `json:"version" example:"2"`
	Definition *json.RawMessage `json:"definition" swaggertype:"object"`
	ValidFrom  *time.Time       `json:"valid_from" example:"2026-07-01T00:00:00Z"`
	ValidTo    *time.Time       `json:"valid_to" example:"2027-07-01T00:00:00Z"`
}

type UserPostRequest struct {
	RoleId   string  `json:"role_id" example:"7a3c2e10-4b1a-4b8e-9f3a-2b6e7c9d1a04"`
	Name     string  `json:"name" example:"Grace Hopper"`
	Bio      *string `json:"bio" example:"Keeps the espresso machines humming and the firmware fresh."`
	Username string  `json:"username" example:"grace.hopper"`
	Password string  `json:"password" example:"BrewMeUp!42"`
}

type UserPatchRequest struct {
	RoleId   *string `json:"role_id" example:"7a3c2e10-4b1a-4b8e-9f3a-2b6e7c9d1a04"`
	Name     *string `json:"name" example:"Grace Hopper"`
	Bio      *string `json:"bio" example:"Keeps the espresso machines humming and the firmware fresh."`
	Username *string `json:"username" example:"grace.hopper"`
}

type UserPasswordPatchRequest struct {
	Password string `json:"password" example:"EvenMoreEspresso!7"`
}

type ApiKeyPostRequest struct {
	UserId    string     `json:"user_id" example:"7a3c2e10-4b1a-4b8e-9f3a-2b6e7c9d1a04"`
	ExpiresAt *time.Time `json:"expires_at" example:"2027-01-01T00:00:00Z"`
}

type ApiKeyRegenerateRequest struct {
	ExpiresAt *time.Time `json:"expires_at" example:"2027-01-01T00:00:00Z"`
}
