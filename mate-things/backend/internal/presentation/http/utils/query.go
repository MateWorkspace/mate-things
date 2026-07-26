package presentationhttputils

import (
	"strconv"
	"strings"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

func QueryUUID(c *echo.Context, field string) (*uuid.UUID, error) {
	value := strings.TrimSpace(c.QueryParam(field))
	if value == "" {
		return nil, nil
	}

	return OptionalUUID(&value, field)
}

func QueryString(c *echo.Context, field string) *string {
	value := strings.TrimSpace(c.QueryParam(field))
	if value == "" {
		return nil
	}

	return &value
}

func QueryInt32(c *echo.Context, field string) (*int32, error) {
	value := strings.TrimSpace(c.QueryParam(field))
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return nil, domainmodels.NewError(field+" must be a valid integer", domainmodels.ErrTypeValidation, err)
	}

	result := int32(parsed)
	return &result, nil
}

func QueryTime(c *echo.Context, field string) (*time.Time, error) {
	value := strings.TrimSpace(c.QueryParam(field))
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, domainmodels.NewError(field+" must be a valid RFC3339 timestamp", domainmodels.ErrTypeValidation, err)
	}

	return &parsed, nil
}
