package presentationhttputils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	"github.com/labstack/echo/v5"
)

func BadRequest(c *echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, presentationhttpresponse.ErrorResponse{
		Error:   "Bad Request",
		Message: message,
	})
}

func Error(c *echo.Context, err error, message string) error {
	if err == nil {
		return nil
	}

	statusCode := http.StatusInternalServerError
	title := "Internal Server Error"

	switch {
	case errors.Is(err, domainmodels.ErrTypeNotFound):
		statusCode = http.StatusNotFound
		title = "Not Found"
	case errors.Is(err, domainmodels.ErrTypeConflict):
		statusCode = http.StatusConflict
		title = "Already Exists"
	case errors.Is(err, domainmodels.ErrTypeBadArgs),
		errors.Is(err, domainmodels.ErrTypeValidation):
		statusCode = http.StatusBadRequest
		title = "Invalid Format"
	case errors.Is(err, domainmodels.ErrTypeBadState):
		statusCode = http.StatusPreconditionFailed
		title = "Invalid State"
	case errors.Is(err, domainmodels.ErrTypeForbidden):
		statusCode = http.StatusForbidden
		title = "Access Denied"
	case errors.Is(err, domainmodels.ErrTypeTokenExpired):
		statusCode = http.StatusUnauthorized
		title = "Session Expired"
	case errors.Is(err, domainmodels.ErrTypeTokenInvalid):
		statusCode = http.StatusUnauthorized
		title = "Invalid Token"
	case errors.Is(err, domainmodels.ErrTypeUnauthorized):
		statusCode = http.StatusUnauthorized
		title = "Unauthorized"
	case errors.Is(err, domainmodels.ErrTypeTimeout):
		statusCode = http.StatusGatewayTimeout
		title = "Request Timeout"
	case errors.Is(err, domainmodels.ErrTypeUnimplemented):
		statusCode = http.StatusNotImplemented
		title = "Not Implemented"
	case errors.Is(err, domainmodels.ErrTypeFailure),
		errors.Is(err, domainmodels.ErrTypeUnknown):
		statusCode = http.StatusInternalServerError
		title = "Internal Server Error"
	}

	return c.JSON(statusCode, presentationhttpresponse.ErrorResponse{
		Error:   title,
		Message: message,
		Details: errorDetails(err),
	})
}

func errorDetails(err error) string {
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
