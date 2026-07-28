package presentationhttpresponse

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type IdResponse struct {
	Id string `json:"id"`
}

type CountResponse struct {
	Count int `json:"count"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details"`
}

type PageResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
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
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedBy *string    `json:"created_by,omitempty"`
	UpdatedBy *string    `json:"updated_by,omitempty"`
	DeletedBy *string    `json:"deleted_by,omitempty"`
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
