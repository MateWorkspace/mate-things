package presentationhttputils

import (
	"encoding/json"
	"strconv"
	"strings"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func RequiredUUID(value string, field string) (uuid.UUID, error) {
	if strings.TrimSpace(value) == "" {
		return uuid.Nil, domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}

	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, domainmodels.NewError(field+" must be a valid UUID", domainmodels.ErrTypeValidation, err)
	}

	return id, nil
}

func OptionalUUID(value *string, field string) (*uuid.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}

	id, err := RequiredUUID(*value, field)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func RequiredString(value string, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}

	return value, nil
}

func RequiredInt32(value string, field string) (int32, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, domainmodels.NewError(field+" must be a valid integer", domainmodels.ErrTypeValidation, err)
	}

	return int32(parsed), nil
}

func OptionalInt32(value *string, field string) (*int32, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}

	parsed, err := RequiredInt32(*value, field)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func RequiredRawJSON(value json.RawMessage, field string) (json.RawMessage, error) {
	if len(value) == 0 {
		return nil, domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}
	if !json.Valid(value) {
		return nil, domainmodels.NewError(field+" must be a valid JSON value", domainmodels.ErrTypeValidation, nil)
	}

	return value, nil
}

func OptionalRawJSON(value *json.RawMessage, field string) (*json.RawMessage, error) {
	if value == nil || len(*value) == 0 {
		return nil, nil
	}
	if !json.Valid(*value) {
		return nil, domainmodels.NewError(field+" must be a valid JSON value", domainmodels.ErrTypeValidation, nil)
	}

	return value, nil
}

func BadPairQuery(left string, right string) error {
	return domainmodels.NewError(left+" and "+right+" are required", domainmodels.ErrTypeValidation, nil)
}
