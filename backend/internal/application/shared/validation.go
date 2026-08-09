// Package applicationshared holds value-validation rules shared across
// usecases: character-set/length/format rules for domain values. The
// presentation layer only validates raw request shape (UUID/int/JSON
// parsing, presence); everything about whether a *value* is acceptable
// belongs here, called from each usecase's Create/UpdateById.
package applicationshared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	nodeNamePattern       = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	snakeCaseNamePattern  = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`)
	firmwareNamePattern   = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
	// A MAC address with the colon separators stripped: 12 hex digits.
	deviceIdPattern = regexp.MustCompile(`^[0-9A-Fa-f]{12}$`)
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

// Node class name: alphanumeric, underscore, and dash. Underscores are part
// of the persisted seed contract (for example, "base_node").

func RequiredNodeClassName(value string, field string) (string, error) {
	return validateNamePattern(value, field, nodeNamePattern, nodeClassNameMinLength, nodeClassNameMaxLength,
		"alphanumeric characters, underscores, and dashes")
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

// Node name follows the same safe persisted contract. Self-registration
// creates names in the form "node_<device_id>".

func RequiredNodeName(value string, field string) (string, error) {
	return validateNamePattern(value, field, nodeNamePattern, nodeNameMinLength, nodeNameMaxLength,
		"alphanumeric characters, underscores, and dashes")
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

// Firmware name: alphanumeric, underscore, dash, dot, no spaces. Dots are
// allowed because firmware names must exactly match the device's own
// registration payload, which embeds a dotted semver-style version string
// (e.g. "mate-espidf-base_v1.0.0-dev.1") - see mate-espidf-base's
// PROJECT_VERSION / registration firmware_name build.

func RequiredFirmwareName(value string, field string) (string, error) {
	return validateNamePattern(value, field, firmwareNamePattern, firmwareNameMinLength, firmwareNameMaxLength,
		"alphanumeric characters, underscores, dashes, and dots")
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

var validLlmProviders = map[domainmodels.LlmProvider]struct{}{
	domainmodels.LlmProviderClaude: {},
	domainmodels.LlmProviderOpenAI: {},
}

func RequiredLlmProvider(value domainmodels.LlmProvider, field string) (domainmodels.LlmProvider, error) {
	if _, ok := validLlmProviders[value]; !ok {
		return "", domainmodels.NewError(fmt.Sprintf("%s must be one of CLAUDE, OPENAI", field), domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}

func RequiredLlmModel(value string, field string) (string, error) {
	if value == "" {
		return "", domainmodels.NewError(fmt.Sprintf("%s is required", field), domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}

func RequiredLlmApiKey(value string, field string) (string, error) {
	if value == "" {
		return "", domainmodels.NewError(fmt.Sprintf("%s is required", field), domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}

func OptionalLlmBaseURL(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	url, err := RequiredURL(*value, field)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

var validPayloadSchemaDefinitionTypes = map[domainmodels.PayloadSchemaDefinitionType]bool{
	domainmodels.PayloadSchemaDefinitionTypeString:       true,
	domainmodels.PayloadSchemaDefinitionTypeFloat:        true,
	domainmodels.PayloadSchemaDefinitionTypeInteger:      true,
	domainmodels.PayloadSchemaDefinitionTypeBoolean:      true,
	domainmodels.PayloadSchemaDefinitionTypeEnum:         true,
	domainmodels.PayloadSchemaDefinitionTypeObject:       true,
	domainmodels.PayloadSchemaDefinitionTypeStringArray:  true,
	domainmodels.PayloadSchemaDefinitionTypeFloatArray:   true,
	domainmodels.PayloadSchemaDefinitionTypeIntegerArray: true,
	domainmodels.PayloadSchemaDefinitionTypeBooleanArray: true,
	domainmodels.PayloadSchemaDefinitionTypeEnumArray:    true,
	domainmodels.PayloadSchemaDefinitionTypeObjectArray:  true,
}

// RequiredPayloadSchemaDefinition validates that a payload schema's
// definition is well-formed JSON that also structurally conforms to
// domainmodels.PayloadSchemaDefinition: every type (including nested
// object properties and array items) is a recognized schema type, and
// enum/[]enum nodes declare at least one option. This catches malformed
// schemas at write time instead of only surfacing at dispatch time when
// a payload is checked against them.
func RequiredPayloadSchemaDefinition(value json.RawMessage, field string) (json.RawMessage, error) {
	if len(value) == 0 {
		return nil, domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}

	var definition domainmodels.PayloadSchemaDefinition
	if err := json.Unmarshal(value, &definition); err != nil {
		return nil, domainmodels.NewError(field+" must be a valid JSON value", domainmodels.ErrTypeValidation, err)
	}
	if err := validatePayloadSchemaDefinitionShape(definition, field); err != nil {
		return nil, err
	}

	return value, nil
}

func OptionalPayloadSchemaDefinition(value *json.RawMessage, field string) (*json.RawMessage, error) {
	if value == nil || len(*value) == 0 {
		return nil, nil
	}

	v, err := RequiredPayloadSchemaDefinition(*value, field)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func validatePayloadSchemaDefinitionShape(definition domainmodels.PayloadSchemaDefinition, field string) error {
	if !validPayloadSchemaDefinitionTypes[definition.Type] {
		return domainmodels.NewError(field+" has an invalid or missing type", domainmodels.ErrTypeValidation, nil)
	}
	if (definition.Type == domainmodels.PayloadSchemaDefinitionTypeEnum || definition.Type == domainmodels.PayloadSchemaDefinitionTypeEnumArray) && len(definition.Options) == 0 {
		return domainmodels.NewError(field+" enum type requires at least one option", domainmodels.ErrTypeValidation, nil)
	}

	for name, property := range definition.Properties {
		if err := validatePayloadSchemaDefinitionShape(property, field+"."+name); err != nil {
			return err
		}
	}
	if definition.Items != nil {
		if err := validatePayloadSchemaDefinitionShape(*definition.Items, field+".items"); err != nil {
			return err
		}
	}

	return nil
}

// RequiredDeviceId validates a node's device_id is a MAC address with the
// colon separators stripped (12 hex digits) - the format the ESP32 firmware
// sends at registration.
func RequiredDeviceId(value string, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}
	if !deviceIdPattern.MatchString(value) {
		return "", domainmodels.NewError(field+" must be a 12-character hex MAC address without separators", domainmodels.ErrTypeValidation, nil)
	}

	return value, nil
}

// ValidateConfigValue checks a config value string against the value_type
// declared for its key in the owning firmware's config schema (string,
// uint32, or bool) - shared by any usecase that accepts a
// (key, value, value_type) config write, whether that write comes from an
// HTTP request or a device's own MQTT registration report.
func ValidateConfigValue(value string, valueType string) error {
	switch valueType {
	case "uint32":
		if _, err := strconv.ParseUint(value, 10, 32); err != nil {
			return domainmodels.NewError("config value must be a valid uint32", domainmodels.ErrTypeValidation, err)
		}
	case "bool":
		if value != "true" && value != "false" {
			return domainmodels.NewError("config value must be \"true\" or \"false\"", domainmodels.ErrTypeValidation, nil)
		}
	case "string":
		// any string value is acceptable
	default:
		return domainmodels.NewError("unrecognized config value_type", domainmodels.ErrTypeFailure, nil)
	}

	return nil
}

// esp32ImageMagicByte is the fixed first byte of every ESP-IDF app/bootloader
// image (esp_image_header_t.magic). Firmware uploads are checked against it
// so an admin uploading the wrong file is rejected at upload time instead of
// only discovered when a device fails to flash it.
const esp32ImageMagicByte = 0xE9

// RequiredFirmwareContent validates that content begins with the ESP32
// image magic byte and returns a reader with the peeked byte restored so
// the full content can still be read/stored afterward.
func RequiredFirmwareContent(content io.Reader, field string) (io.Reader, error) {
	if content == nil {
		return nil, domainmodels.NewError(field+" is required", domainmodels.ErrTypeValidation, nil)
	}

	header := make([]byte, 1)
	if _, err := io.ReadFull(content, header); err != nil {
		return nil, domainmodels.NewError(field+" must be a valid ESP32 firmware image", domainmodels.ErrTypeValidation, err)
	}
	if header[0] != esp32ImageMagicByte {
		return nil, domainmodels.NewError(field+" must be a valid ESP32 firmware image", domainmodels.ErrTypeValidation, nil)
	}

	return io.MultiReader(bytes.NewReader(header), content), nil
}

var validInfraredRecordingStates = map[string]struct{}{
	domainmodels.InfraredRecordingStateDraft:               {},
	domainmodels.InfraredRecordingStateCasesGenerating:     {},
	domainmodels.InfraredRecordingStateRecording:           {},
	domainmodels.InfraredRecordingStateAnalyzing:           {},
	domainmodels.InfraredRecordingStateFunctionGenerating:  {},
	domainmodels.InfraredRecordingStateTestCasesGenerating: {},
	domainmodels.InfraredRecordingStateTesting:             {},
	domainmodels.InfraredRecordingStateCompleted:           {},
	domainmodels.InfraredRecordingStateFailed:              {},
}

func RequiredInfraredRecordingState(value string, field string) (string, error) {
	if _, ok := validInfraredRecordingStates[value]; !ok {
		return "", domainmodels.NewError(fmt.Sprintf("%s is not a recognized recording state", field), domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}

func RequiredInfraredBrand(value string, field string) (string, error) {
	if value == "" {
		return "", domainmodels.NewError(fmt.Sprintf("%s is required", field), domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}

func RequiredInfraredModel(value string, field string) (string, error) {
	if value == "" {
		return "", domainmodels.NewError(fmt.Sprintf("%s is required", field), domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}
