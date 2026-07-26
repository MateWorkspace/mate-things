package presentationhttputils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	presentationhttpresponse "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/response"
	"github.com/labstack/echo/v5"
)

func BadRequest(c *echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, presentationhttpresponse.ErrorResponse{
		Code:    "bad_request",
		Message: message,
	})
}

func Error(c *echo.Context, err error) error {
	if err == nil {
		return nil
	}

	statusCode := http.StatusInternalServerError
	code := "unknown"

	switch {
	case errors.Is(err, domainmodels.ErrTypeNotFound):
		statusCode = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, domainmodels.ErrTypeConflict):
		statusCode = http.StatusConflict
		code = "conflict"
	case errors.Is(err, domainmodels.ErrTypeBadArgs),
		errors.Is(err, domainmodels.ErrTypeValidation):
		statusCode = http.StatusBadRequest
		code = "validation"
	case errors.Is(err, domainmodels.ErrTypeBadState):
		statusCode = http.StatusPreconditionFailed
		code = "bad_state"
	case errors.Is(err, domainmodels.ErrTypeUnauthorized),
		errors.Is(err, domainmodels.ErrTypeTokenExpired),
		errors.Is(err, domainmodels.ErrTypeTokenInvalid):
		statusCode = http.StatusUnauthorized
		code = "unauthorized"
	case errors.Is(err, domainmodels.ErrTypeTimeout):
		statusCode = http.StatusGatewayTimeout
		code = "timeout"
	case errors.Is(err, domainmodels.ErrTypeUnimplemented):
		statusCode = http.StatusNotImplemented
		code = "unimplemented"
	case errors.Is(err, domainmodels.ErrTypeFailure),
		errors.Is(err, domainmodels.ErrTypeUnknown):
		statusCode = http.StatusInternalServerError
		code = "unknown"
	}

	return c.JSON(statusCode, presentationhttpresponse.ErrorResponse{
		Code:    code,
		Message: errorMessage(err),
	})
}

func errorMessage(err error) string {
	var domainErr *domainmodels.Error
	if errors.As(err, &domainErr) && strings.TrimSpace(domainErr.Message) != "" {
		return domainErr.Message
	}

	return err.Error()
}

func MissingResponse(name string) error {
	return domainmodels.NewError(
		fmt.Sprintf("%s is missing from usecase response", name),
		domainmodels.ErrTypeUnknown,
		nil,
	)
}
