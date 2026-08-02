package applicationnodeconfigvalue

import (
	"errors"
	"testing"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

func TestValidateConfigValuePreservesStringDomainAndRejectsInvalidTypedValues(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		valueType string
		wantError bool
	}{
		{name: "empty string", value: "", valueType: "string"},
		{name: "whitespace string", value: "  ", valueType: "string"},
		{name: "uint32 negative", value: "-1", valueType: "uint32", wantError: true},
		{name: "uint32 fractional", value: "1.5", valueType: "uint32", wantError: true},
		{name: "uint32 overflow", value: "4294967296", valueType: "uint32", wantError: true},
		{name: "bool invalid", value: "yes", valueType: "bool", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := applicationshared.ValidateConfigValue(test.value, test.valueType)

			if test.wantError && !errors.Is(err, domainmodels.ErrTypeValidation) {
				t.Fatalf("ValidateConfigValue() error = %v, want validation error", err)
			}
			if !test.wantError && err != nil {
				t.Fatalf("ValidateConfigValue() error = %v, want nil", err)
			}
		})
	}
}
