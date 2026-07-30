package domainmodels

import (
	"errors"
	"fmt"
)

type ErrorType error

var (
	ErrTypeNotFound      = errors.New("NOT_FOUND")
	ErrTypeBadArgs       = errors.New("BAD_ARGS")
	ErrTypeConflict      = errors.New("CONFLICT")
	ErrTypeBadState      = errors.New("BAD_STATE")
	ErrTypeValidation    = errors.New("VALIDATION")
	ErrTypeForbidden     = errors.New("FORBIDDEN")
	ErrTypeUnauthorized  = errors.New("UNAUTHORIZED")
	ErrTypeTokenExpired  = errors.New("TOKEN_EXPIRED")
	ErrTypeTokenInvalid  = errors.New("TOKEN_INVALID")
	ErrTypeTimeout       = errors.New("TIMEOUT")
	ErrTypeUnimplemented = errors.New("UNIMPLEMENTED")
	ErrTypeFailure       = errors.New("FAILURE")
	ErrTypeUnknown       = errors.New("UNKNOWN")

	ErrTypeUsernameExists             = errors.New("USERNAME_EXISTS")
	ErrTypeRoleNameExists             = errors.New("ROLE_NAME_EXISTS")
	ErrTypePermissionNameExists       = errors.New("PERMISSION_NAME_EXISTS")
	ErrTypeActionNameExists           = errors.New("ACTION_NAME_EXISTS")
	ErrTypeNodeClassNameExists        = errors.New("NODE_CLASS_NAME_EXISTS")
	ErrTypeFirmwareNameExists         = errors.New("FIRMWARE_NAME_EXISTS")
	ErrTypeNodeDeviceIdExists         = errors.New("NODE_DEVICE_ID_EXISTS")
	ErrTypePayloadSchemaVersionExists = errors.New("PAYLOAD_SCHEMA_VERSION_EXISTS")
	ErrTypeRolePermissionExists       = errors.New("ROLE_PERMISSION_EXISTS")
	ErrTypeFirmwareConfigKeyExists    = errors.New("FIRMWARE_CONFIG_KEY_EXISTS")
	ErrTypeNodeConfigKeyExists        = errors.New("NODE_CONFIG_KEY_EXISTS")
)

type Error struct {
	Message string
	Type    ErrorType
	Source  error
}

func NewError(message string, errType ErrorType, source error) error {
	return &Error{
		Message: message,
		Type:    errType,
		Source:  source,
	}
}

func (e *Error) Error() string {
	if e.Source != nil {
		return fmt.Sprintf("[%v] %s: %v", e.Type, e.Message, e.Source)
	}
	return fmt.Sprintf("[%v] %s", e.Type, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Source
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.Type, target)
}
