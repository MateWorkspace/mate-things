package infrastructureutilitypayloadschemavalidator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type validatorImpl struct{}

func NewValidatorImpl() domaincontractsutility.PayloadSchemaValidator {
	return &validatorImpl{}
}

func (v *validatorImpl) Validate(ctx context.Context, definition json.RawMessage, payload json.RawMessage) error {
	if err := ctx.Err(); err != nil {
		return domainmodels.NewError("payload schema validation cancelled", domainmodels.ErrTypeTimeout, err)
	}

	var schema domainmodels.PayloadSchemaDefinition
	if err := json.Unmarshal(definition, &schema); err != nil {
		return domainmodels.NewError("payload schema definition is invalid", domainmodels.ErrTypeValidation, err)
	}

	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		return domainmodels.NewError("payload is invalid json", domainmodels.ErrTypeValidation, err)
	}

	if err := validateValue(ctx, "$", schema, value); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return domainmodels.NewError("payload schema validation cancelled", domainmodels.ErrTypeTimeout, ctxErr)
		}
		return domainmodels.NewError("payload does not match schema definition", domainmodels.ErrTypeValidation, err)
	}

	return nil
}

func validateValue(ctx context.Context, path string, schema domainmodels.PayloadSchemaDefinition, value any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	switch schema.Type {
	case domainmodels.PayloadSchemaDefinitionTypeString:
		return validateString(path, schema, value)
	case domainmodels.PayloadSchemaDefinitionTypeFloat:
		return validateFloat(path, schema, value)
	case domainmodels.PayloadSchemaDefinitionTypeInteger:
		return validateInteger(path, schema, value)
	case domainmodels.PayloadSchemaDefinitionTypeBoolean:
		return validateBoolean(path, value)
	case domainmodels.PayloadSchemaDefinitionTypeEnum:
		return validateEnum(path, schema, value)
	case domainmodels.PayloadSchemaDefinitionTypeObject:
		return validateObject(ctx, path, schema, value)
	case domainmodels.PayloadSchemaDefinitionTypeStringArray:
		return validateArray(ctx, path, schema, arrayItemSchema(schema, domainmodels.PayloadSchemaDefinitionTypeString), value)
	case domainmodels.PayloadSchemaDefinitionTypeFloatArray:
		return validateArray(ctx, path, schema, arrayItemSchema(schema, domainmodels.PayloadSchemaDefinitionTypeFloat), value)
	case domainmodels.PayloadSchemaDefinitionTypeIntegerArray:
		return validateArray(ctx, path, schema, arrayItemSchema(schema, domainmodels.PayloadSchemaDefinitionTypeInteger), value)
	case domainmodels.PayloadSchemaDefinitionTypeBooleanArray:
		return validateArray(ctx, path, schema, arrayItemSchema(schema, domainmodels.PayloadSchemaDefinitionTypeBoolean), value)
	case domainmodels.PayloadSchemaDefinitionTypeEnumArray:
		return validateArray(ctx, path, schema, arrayItemSchema(schema, domainmodels.PayloadSchemaDefinitionTypeEnum), value)
	case domainmodels.PayloadSchemaDefinitionTypeObjectArray:
		return validateArray(ctx, path, schema, arrayItemSchema(schema, domainmodels.PayloadSchemaDefinitionTypeObject), value)
	default:
		return fmt.Errorf("%s has unsupported schema type %q", path, schema.Type)
	}
}

func validateString(path string, schema domainmodels.PayloadSchemaDefinition, value any) error {
	text, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s must be string", path)
	}

	length := int64(len(text))
	if schema.MinimumLength != nil && length < *schema.MinimumLength {
		return fmt.Errorf("%s length must be at least %d", path, *schema.MinimumLength)
	}
	if schema.MaximumLength != nil && length > *schema.MaximumLength {
		return fmt.Errorf("%s length must be at most %d", path, *schema.MaximumLength)
	}

	return nil
}

func validateFloat(path string, schema domainmodels.PayloadSchemaDefinition, value any) error {
	number, ok := value.(float64)
	if !ok {
		return fmt.Errorf("%s must be number", path)
	}

	return validateNumberRange(path, schema, number)
}

func validateInteger(path string, schema domainmodels.PayloadSchemaDefinition, value any) error {
	number, ok := value.(float64)
	if !ok {
		return fmt.Errorf("%s must be integer", path)
	}
	if math.Trunc(number) != number {
		return fmt.Errorf("%s must be integer", path)
	}

	return validateNumberRange(path, schema, number)
}

func validateBoolean(path string, value any) error {
	if _, ok := value.(bool); !ok {
		return fmt.Errorf("%s must be boolean", path)
	}

	return nil
}

func validateEnum(path string, schema domainmodels.PayloadSchemaDefinition, value any) error {
	text, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s must be enum string", path)
	}
	if len(schema.Options) == 0 {
		return fmt.Errorf("%s enum options are required", path)
	}

	for _, option := range schema.Options {
		if text == option {
			return nil
		}
	}

	return fmt.Errorf("%s must be one of schema enum options", path)
}

func validateObject(ctx context.Context, path string, schema domainmodels.PayloadSchemaDefinition, value any) error {
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be object", path)
	}

	for _, field := range schema.Required {
		if _, ok := object[field]; !ok {
			return fmt.Errorf("%s.%s is required", path, field)
		}
	}

	for field, propertySchema := range schema.Properties {
		propertyValue, ok := object[field]
		if !ok {
			continue
		}
		if err := validateValue(ctx, path+"."+field, propertySchema, propertyValue); err != nil {
			return err
		}
	}

	return nil
}

func validateArray(
	ctx context.Context,
	path string,
	schema domainmodels.PayloadSchemaDefinition,
	itemSchema domainmodels.PayloadSchemaDefinition,
	value any,
) error {
	items, ok := value.([]any)
	if !ok {
		return fmt.Errorf("%s must be array", path)
	}

	length := int64(len(items))
	if schema.MinimumItem != nil && length < *schema.MinimumItem {
		return fmt.Errorf("%s item count must be at least %d", path, *schema.MinimumItem)
	}
	if schema.MaximumItem != nil && length > *schema.MaximumItem {
		return fmt.Errorf("%s item count must be at most %d", path, *schema.MaximumItem)
	}

	for i, item := range items {
		if err := validateValue(ctx, fmt.Sprintf("%s[%d]", path, i), itemSchema, item); err != nil {
			return err
		}
	}

	return nil
}

func validateNumberRange(path string, schema domainmodels.PayloadSchemaDefinition, number float64) error {
	if schema.Minimum != nil && number < *schema.Minimum {
		return fmt.Errorf("%s must be at least %v", path, *schema.Minimum)
	}
	if schema.Maximum != nil && number > *schema.Maximum {
		return fmt.Errorf("%s must be at most %v", path, *schema.Maximum)
	}

	return nil
}

func arrayItemSchema(
	schema domainmodels.PayloadSchemaDefinition,
	defaultType domainmodels.PayloadSchemaDefinitionType,
) domainmodels.PayloadSchemaDefinition {
	item := domainmodels.PayloadSchemaDefinition{
		Type:          defaultType,
		Required:      schema.Required,
		Options:       schema.Options,
		Unit:          schema.Unit,
		Minimum:       schema.Minimum,
		Maximum:       schema.Maximum,
		MinimumLength: schema.MinimumLength,
		MaximumLength: schema.MaximumLength,
		Properties:    schema.Properties,
	}

	if schema.Items != nil {
		item = *schema.Items
		if item.Type == "" {
			item.Type = defaultType
		}
	}

	return item
}
