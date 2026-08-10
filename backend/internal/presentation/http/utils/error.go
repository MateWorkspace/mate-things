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

type errorMapping struct {
	status  int
	title   string
	message string // non-empty: always used, overriding the domain error's own message
}

var errorMappings = []struct {
	errType domainmodels.ErrorType
	mapping errorMapping
}{
	{domainmodels.ErrTypeNotFound, errorMapping{http.StatusNotFound, "Not Found", ""}},
	{domainmodels.ErrTypeUsernameExists, errorMapping{http.StatusConflict, "Already Exists", "This username is already taken."}},
	{domainmodels.ErrTypeRoleNameExists, errorMapping{http.StatusConflict, "Already Exists", "A role with this name already exists."}},
	{domainmodels.ErrTypePermissionNameExists, errorMapping{http.StatusConflict, "Already Exists", "A permission with this name already exists."}},
	{domainmodels.ErrTypeActionNameExists, errorMapping{http.StatusConflict, "Already Exists", "An action with this name already exists."}},
	{domainmodels.ErrTypeNodeClassNameExists, errorMapping{http.StatusConflict, "Already Exists", "A node class with this name already exists."}},
	{domainmodels.ErrTypeFirmwareNameExists, errorMapping{http.StatusConflict, "Already Exists", "A firmware with this name already exists."}},
	{domainmodels.ErrTypeNodeDeviceIdExists, errorMapping{http.StatusConflict, "Already Exists", "A node with this device ID is already registered."}},
	{domainmodels.ErrTypePayloadSchemaVersionExists, errorMapping{http.StatusConflict, "Already Exists", "This payload schema name and version already exists."}},
	{domainmodels.ErrTypeRolePermissionExists, errorMapping{http.StatusConflict, "Already Exists", "This permission is already assigned to the role."}},
	{domainmodels.ErrTypeFirmwareConfigKeyExists, errorMapping{http.StatusConflict, "Already Exists", "This config key already exists for the firmware."}},
	{domainmodels.ErrTypeNodeConfigKeyExists, errorMapping{http.StatusConflict, "Already Exists", "This config key already exists for the node."}},
	{domainmodels.ErrTypeApiKeyUserExists, errorMapping{http.StatusConflict, "Already Exists", "This user already has an API key. Regenerate it instead."}},
	{domainmodels.ErrTypeConflict, errorMapping{http.StatusConflict, "Already Exists", ""}},
	{domainmodels.ErrTypeBroadcastListenerLimitReached, errorMapping{http.StatusConflict, "Already Being Watched", "This recording session already has an active listener."}},
	{domainmodels.ErrTypeBadArgs, errorMapping{http.StatusBadRequest, "Invalid Format", ""}},
	{domainmodels.ErrTypeValidation, errorMapping{http.StatusBadRequest, "Invalid Format", ""}},
	{domainmodels.ErrTypeBadState, errorMapping{http.StatusPreconditionFailed, "Invalid State", ""}},
	{domainmodels.ErrTypeForbidden, errorMapping{http.StatusForbidden, "Access Denied", ""}},
	{domainmodels.ErrTypeUnauthorized, errorMapping{http.StatusUnauthorized, "Unauthorized", ""}},
	{domainmodels.ErrTypeTokenExpired, errorMapping{http.StatusUnauthorized, "Session Expired", "Your session has expired. Please sign in again."}},
	{domainmodels.ErrTypeTokenInvalid, errorMapping{http.StatusUnauthorized, "Invalid Token", "Your session is no longer valid. Please sign in again."}},
	{domainmodels.ErrTypeTimeout, errorMapping{http.StatusGatewayTimeout, "Request Timeout", "The request took too long. Please try again."}},
	{domainmodels.ErrTypeUnimplemented, errorMapping{http.StatusNotImplemented, "Not Implemented", "This feature isn't available yet."}},
	{domainmodels.ErrTypeFailure, errorMapping{http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."}},
	{domainmodels.ErrTypeUnknown, errorMapping{http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."}},
}

const genericServerErrorMessage = "Something went wrong on our end. Please try again later."

func Error(c *echo.Context, err error) error {
	if err == nil {
		return nil
	}

	for _, m := range errorMappings {
		if errors.Is(err, m.errType) {
			message := m.mapping.message
			if message == "" {
				message = domainMessage(err)
			}
			return c.JSON(m.mapping.status, presentationhttpresponse.ErrorResponse{
				Error:   m.mapping.title,
				Message: message,
			})
		}
	}

	return c.JSON(http.StatusInternalServerError, presentationhttpresponse.ErrorResponse{
		Error:   "Internal Server Error",
		Message: genericServerErrorMessage,
	})
}

func domainMessage(err error) string {
	var domainErr *domainmodels.Error
	if errors.As(err, &domainErr) && strings.TrimSpace(domainErr.Message) != "" {
		return domainErr.Message
	}
	return genericServerErrorMessage
}

func MissingResponse(name string) error {
	return domainmodels.NewError(
		fmt.Sprintf("%s is missing from usecase response", name),
		domainmodels.ErrTypeUnknown,
		nil,
	)
}
