package applicationshared

import (
	"errors"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

func TestRequiredLlmProviderRejectsUnknownValue(t *testing.T) {
	_, err := RequiredLlmProvider(domainmodels.LlmProvider("GROK"), "provider")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredLlmProvider() error = %v, want validation error", err)
	}
}

func TestRequiredLlmProviderAcceptsKnownValues(t *testing.T) {
	for _, provider := range []domainmodels.LlmProvider{domainmodels.LlmProviderClaude, domainmodels.LlmProviderOpenAI, domainmodels.LlmProviderGemini} {
		got, err := RequiredLlmProvider(provider, "provider")
		if err != nil {
			t.Fatalf("RequiredLlmProvider(%q) error = %v, want nil", provider, err)
		}
		if got != provider {
			t.Fatalf("RequiredLlmProvider(%q) = %q, want %q", provider, got, provider)
		}
	}
}

func TestRequiredLlmModelRejectsEmpty(t *testing.T) {
	_, err := RequiredLlmModel("", "model")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredLlmModel(\"\") error = %v, want validation error", err)
	}
}

func TestRequiredLlmApiKeyRejectsEmpty(t *testing.T) {
	_, err := RequiredLlmApiKey("", "api_key")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredLlmApiKey(\"\") error = %v, want validation error", err)
	}
}

func TestOptionalLlmBaseURLRejectsMalformedURL(t *testing.T) {
	value := "not a url"
	_, err := OptionalLlmBaseURL(&value, "base_url")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("OptionalLlmBaseURL(%q) error = %v, want validation error", value, err)
	}
}

func TestOptionalLlmBaseURLAllowsNil(t *testing.T) {
	got, err := OptionalLlmBaseURL(nil, "base_url")
	if err != nil {
		t.Fatalf("OptionalLlmBaseURL(nil) error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("OptionalLlmBaseURL(nil) = %v, want nil", got)
	}
}

func TestOptionalLlmApiKeyTreatsNilAndEmptyAsUnset(t *testing.T) {
	if got := OptionalLlmApiKey(nil); got != nil {
		t.Fatalf("OptionalLlmApiKey(nil) = %v, want nil", got)
	}
	empty := ""
	if got := OptionalLlmApiKey(&empty); got != nil {
		t.Fatalf("OptionalLlmApiKey(\"\") = %v, want nil", got)
	}
}

func TestOptionalLlmApiKeyReturnsProvidedValue(t *testing.T) {
	value := "sk-ant-real-key"
	got := OptionalLlmApiKey(&value)
	if got == nil || *got != value {
		t.Fatalf("OptionalLlmApiKey(%q) = %v, want %q", value, got, value)
	}
}
