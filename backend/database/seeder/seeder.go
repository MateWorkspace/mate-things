package seeder

import (
	"embed"
	"encoding/json"
)

//go:embed permission.json role.json node_class.json payload_schema.json action.json user.json infrared_device_type.json infrared_state.json
var files embed.FS

type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsDefault   bool     `json:"is_default"`
	Permissions []string `json:"permissions"`
}

type NodeClass struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PayloadSchema struct {
	Name       string          `json:"name"`
	Version    int32           `json:"version"`
	Definition json.RawMessage `json:"definition"`
}

type Action struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	NodeClassNames       []string `json:"node_class_names"`
	PayloadSchemaName    string   `json:"payload_schema_name"`
	PayloadSchemaVersion int32    `json:"payload_schema_version"`
}

type User struct {
	RoleName string  `json:"role_name"`
	Name     string  `json:"name"`
	Bio      *string `json:"bio,omitempty"`
	Username string  `json:"username"`
	Password string  `json:"password"`
}

type InfraredDeviceType struct {
	Name string `json:"name"`
}

type InfraredState struct {
	DeviceTypeName string `json:"device_type_name"`
	Name           string `json:"name"`
	Type           string `json:"type"`
}

type Data struct {
	Permissions         []Permission
	Roles               []Role
	NodeClasses         []NodeClass
	PayloadSchemas      []PayloadSchema
	Actions             []Action
	Users               []User
	InfraredDeviceTypes []InfraredDeviceType
	InfraredStates      []InfraredState
}

func Load() (Data, error) {
	var data Data

	targets := map[string]any{
		"permission.json":           &data.Permissions,
		"role.json":                 &data.Roles,
		"node_class.json":           &data.NodeClasses,
		"payload_schema.json":       &data.PayloadSchemas,
		"action.json":               &data.Actions,
		"user.json":                 &data.Users,
		"infrared_device_type.json": &data.InfraredDeviceTypes,
		"infrared_state.json":       &data.InfraredStates,
	}

	for name, out := range targets {
		raw, err := files.ReadFile(name)
		if err != nil {
			return Data{}, err
		}
		if err := json.Unmarshal(raw, out); err != nil {
			return Data{}, err
		}
	}

	return data, nil
}
