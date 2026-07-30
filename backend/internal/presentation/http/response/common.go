package presentationhttpresponse

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type IdResponse struct {
	Id string `json:"id" example:"6d9e2f5a-8b1c-4d3e-9f6a-2c5d8e1f4b07"`
}

type CountResponse struct {
	Count int `json:"count" example:"12"`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid Format"`
	Message string `json:"message" example:"Please select a valid node class."`
}

type PageResponse struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"20"`
	TotalItems int `json:"total_items" example:"87"`
}

type PageDataResponse[T any] struct {
	Data []T          `json:"data"`
	Page PageResponse `json:"page"`
}

type CountDataResponse[T any] struct {
	Data       []T `json:"data"`
	TotalItems int `json:"total_items"`
}

type AuditResponse struct {
	CreatedAt time.Time  `json:"created_at" example:"2026-06-15T09:30:00Z"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" example:"2026-07-01T14:05:00Z"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" example:"2026-07-20T11:00:00Z"`
	CreatedBy *string    `json:"created_by,omitempty" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	UpdatedBy *string    `json:"updated_by,omitempty" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	DeletedBy *string    `json:"deleted_by,omitempty" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
}

func Audit(
	createdAt time.Time,
	updatedAt *time.Time,
	deletedAt *time.Time,
	createdBy *uuid.UUID,
	updatedBy *uuid.UUID,
	deletedBy *uuid.UUID,
) AuditResponse {
	return AuditResponse{
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
		CreatedBy: UUIDPtrString(createdBy),
		UpdatedBy: UUIDPtrString(updatedBy),
		DeletedBy: UUIDPtrString(deletedBy),
	}
}

func UUIDString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

func UUIDPtrString(id *uuid.UUID) *string {
	if id == nil || *id == uuid.Nil {
		return nil
	}

	value := id.String()
	return &value
}

func NormalizeJSON(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage(`{}`)
	}

	return value
}
