package domainusecasespreferences

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type Update interface {
	Action(ctx context.Context, request UpdateActionPreferencesRequest) error
	Firmware(ctx context.Context, request UpdateFirmwarePreferencesRequest) error
	Node(ctx context.Context, request UpdateNodePreferencesRequest) error
	NodeClass(ctx context.Context, request UpdateNodeClassPreferencesRequest) error
	PayloadSchema(ctx context.Context, request UpdatePayloadSchemaPreferencesRequest) error
	Permission(ctx context.Context, request UpdatePermissionPreferencesRequest) error
	Role(ctx context.Context, request UpdateRolePreferencesRequest) error
	User(ctx context.Context, request UpdateUserPreferencesRequest) error
}

type UpdateActionPreferencesRequest struct {
	Id          uuid.UUID
	Preferences json.RawMessage
	UpdatedBy   *uuid.UUID
}

type UpdateFirmwarePreferencesRequest struct {
	Id          uuid.UUID
	Preferences json.RawMessage
	UpdatedBy   *uuid.UUID
}

type UpdateNodePreferencesRequest struct {
	Id          uuid.UUID
	Preferences json.RawMessage
	UpdatedBy   *uuid.UUID
}

type UpdateNodeClassPreferencesRequest struct {
	Id          uuid.UUID
	Preferences json.RawMessage
	UpdatedBy   *uuid.UUID
}

type UpdatePayloadSchemaPreferencesRequest struct {
	Id          uuid.UUID
	Preferences json.RawMessage
	UpdatedBy   *uuid.UUID
}

type UpdatePermissionPreferencesRequest struct {
	Id          uuid.UUID
	Preferences json.RawMessage
	UpdatedBy   *uuid.UUID
}

type UpdateRolePreferencesRequest struct {
	Id          uuid.UUID
	Preferences json.RawMessage
	UpdatedBy   *uuid.UUID
}

type UpdateUserPreferencesRequest struct {
	Id          uuid.UUID
	Preferences json.RawMessage
	UpdatedBy   *uuid.UUID
}
