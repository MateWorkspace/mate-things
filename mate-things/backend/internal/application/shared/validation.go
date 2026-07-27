// Package applicationshared holds value-validation rules shared across
// usecases: character-set/length/format rules for domain values. The
// presentation layer only validates raw request shape (UUID/int/JSON
// parsing, presence); everything about whether a *value* is acceptable
// belongs here, called from each usecase's Create/UpdateById.
package applicationshared

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

var (
	permissionNamePattern = regexp.MustCompile(`^[A-Za-z0-9/_:-]+$`)
	roleNamePattern       = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	usernamePattern       = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	passwordPattern       = regexp.MustCompile(`^[\x21-\x7E]+$`)
	personNamePattern     = regexp.MustCompile(`^[A-Za-z0-9' -]+$`)
	alphanumericPattern   = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	snakeCaseNamePattern  = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`)
	firmwareNamePattern   = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

const (
	permissionNameMinLength = 3
	permissionNameMaxLength = 128

	roleNameMinLength = 3
	roleNameMaxLength = 128

	usernameMinLength = 3
	usernameMaxLength = 128

	// bcrypt.GenerateFromPassword hard-rejects input longer than 72 bytes;
	// capped here below the nominal 128 so a valid password can never
	// crash hashing.
	passwordMinLength = 8
	passwordMaxLength = 72

	personNameMinLength = 1
	personNameMaxLength = 128

	nodeClassNameMinLength = 1
	nodeClassNameMaxLength = 128

	nodeNameMinLength = 1
	nodeNameMaxLength = 128

	payloadSchemaNameMinLength = 1
	payloadSchemaNameMaxLength = 128

	actionNameMinLength = 1
	actionNameMaxLength = 128

	firmwareNameMinLength = 1
	firmwareNameMaxLength = 128
)

func validateNamePattern(
	value string,
	field string,
	pattern *regexp.Regexp,
	minLength int,
	maxLength int,
	charsetDescription string,
) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}
	if len(value) < minLength || len(value) > maxLength {
		message := field + " must be between " + strconv.Itoa(minLength) + " and " + strconv.Itoa(maxLength) + " characters"
		return "", domainmodels.NewError(message, domainmodels.ErrTypeValidation, nil)
	}
	if !pattern.MatchString(value) {
		return "", domainmodels.NewError(field+" may only contain "+charsetDescription, domainmodels.ErrTypeValidation, nil)
	}

	return value, nil
}

// Permission name: alphanumeric, slash, underscore, dash, and colon (the
// established "resource:action" seed-data convention), no spaces.

func RequiredPermissionName(value string, field string) (string, error) {
	return validateNamePattern(value, field, permissionNamePattern, permissionNameMinLength, permissionNameMaxLength,
		"alphanumeric characters, slashes, underscores, dashes, and colons")
}

func OptionalPermissionName(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredPermissionName(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Role name: alphanumeric, underscore, dash, no spaces.

func RequiredRoleName(value string, field string) (string, error) {
	return validateNamePattern(value, field, roleNamePattern, roleNameMinLength, roleNameMaxLength,
		"alphanumeric characters, underscores, and dashes")
}

func OptionalRoleName(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredRoleName(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Username: alphanumeric, underscore, dash, no spaces.

func RequiredUsername(value string, field string) (string, error) {
	return validateNamePattern(value, field, usernamePattern, usernameMinLength, usernameMaxLength,
		"alphanumeric characters, underscores, and dashes")
}

func OptionalUsername(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredUsername(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Password: alphanumeric and regular printable symbols, no spaces/control
// characters. No Optional variant: every request that carries a password
// field requires it (create, admin reset, self-service change).

func RequiredPassword(value string, field string) (string, error) {
	if value == "" {
		return "", domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}
	if len(value) < passwordMinLength || len(value) > passwordMaxLength {
		message := field + " must be between " + strconv.Itoa(passwordMinLength) + " and " + strconv.Itoa(passwordMaxLength) + " characters"
		return "", domainmodels.NewError(message, domainmodels.ErrTypeValidation, nil)
	}
	if !passwordPattern.MatchString(value) {
		return "", domainmodels.NewError(field+" may only contain alphanumeric characters and regular symbols, no spaces", domainmodels.ErrTypeValidation, nil)
	}

	return value, nil
}

// Person name (a user's display name): alphanumeric, space, dash, and
// apostrophe - a regular human-name typeset.

func RequiredPersonName(value string, field string) (string, error) {
	return validateNamePattern(value, field, personNamePattern, personNameMinLength, personNameMaxLength,
		"alphanumeric characters, spaces, dashes, and apostrophes")
}

func OptionalPersonName(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredPersonName(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Node class name: alphanumeric only.

func RequiredNodeClassName(value string, field string) (string, error) {
	return validateNamePattern(value, field, alphanumericPattern, nodeClassNameMinLength, nodeClassNameMaxLength,
		"alphanumeric characters")
}

func OptionalNodeClassName(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredNodeClassName(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Node name: alphanumeric only, same as node class.

func RequiredNodeName(value string, field string) (string, error) {
	return validateNamePattern(value, field, alphanumericPattern, nodeNameMinLength, nodeNameMaxLength,
		"alphanumeric characters")
}

func OptionalNodeName(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredNodeName(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// snake_case name, shared by payload_schema names and action names
// (including an action's payload_schema_name reference field).

func RequiredSnakeCaseName(value string, field string) (string, error) {
	return validateNamePattern(value, field, snakeCaseNamePattern, payloadSchemaNameMinLength, payloadSchemaNameMaxLength,
		"lowercase alphanumeric characters and underscores, in snake_case")
}

func OptionalSnakeCaseName(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredSnakeCaseName(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Action name: snake_case, same as payload schema names.

func RequiredActionName(value string, field string) (string, error) {
	return validateNamePattern(value, field, snakeCaseNamePattern, actionNameMinLength, actionNameMaxLength,
		"lowercase alphanumeric characters and underscores, in snake_case")
}

func OptionalActionName(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredActionName(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Firmware name: alphanumeric, underscore, dash, no spaces.

func RequiredFirmwareName(value string, field string) (string, error) {
	return validateNamePattern(value, field, firmwareNamePattern, firmwareNameMinLength, firmwareNameMaxLength,
		"alphanumeric characters, underscores, and dashes")
}

func OptionalFirmwareName(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredFirmwareName(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Positive integer version numbers (payload_schema.version,
// action.payload_schema_version).

func RequiredPositiveVersion(value int32, field string) (int32, error) {
	if value < 1 {
		return 0, domainmodels.NewError(field+" must be a positive integer", domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}

func OptionalPositiveVersion(value *int32, field string) (*int32, error) {
	if value == nil {
		return nil, nil
	}
	v, err := RequiredPositiveVersion(*value, field)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ValidateTimeWindow rejects a validTo that doesn't come after validFrom.
// Either bound may be nil (open-ended); only checked when both are present.
func ValidateTimeWindow(validFrom *time.Time, validTo *time.Time) error {
	if validFrom != nil && validTo != nil && !validTo.After(*validFrom) {
		return domainmodels.NewError("valid_to must be after valid_from", domainmodels.ErrTypeValidation, nil)
	}
	return nil
}

// RequiredURL validates a value looks like an absolute URL (scheme+host).
func RequiredURL(value string, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", domainmodels.NewError(field+" must be a valid URL", domainmodels.ErrTypeValidation, err)
	}

	return value, nil
}
