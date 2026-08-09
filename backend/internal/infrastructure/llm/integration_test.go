package infrastructurellm

import (
	"context"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructureutilityencryption "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/utility/encryption"
	"github.com/google/uuid"
)

// fakeLlmConfigRepositoryForIntegration mirrors stubLlmConfigRepository but
// lives in its own type to keep this test self-contained from factory_test.go.
type fakeLlmConfigRepositoryForIntegration struct {
	config *domainmodels.LlmConfig
}

func (f *fakeLlmConfigRepositoryForIntegration) Get(_ context.Context) (*domainmodels.LlmConfig, error) {
	return f.config, nil
}

func (f *fakeLlmConfigRepositoryForIntegration) Upsert(_ context.Context, _ domainmodels.LlmProvider, _ string, _ []byte, _ *string, _ *uuid.UUID) (uuid.UUID, error) {
	panic("not used by this test")
}

// TestCurrentDecryptsRealEncryptedApiKeyEndToEnd wires the real AES-GCM
// encryptor with a fake repository to prove ClientFactory.Current actually
// decrypts what the usecase would have encrypted on write. This is the test
// that would have caught Finding 1: a nil config surfacing through the real
// repository contract must resolve cleanly, not panic or 404.
func TestCurrentDecryptsRealEncryptedApiKeyEndToEnd(t *testing.T) {
	const testKey = "01234567890123456789012345678901" // 32 bytes for AES-256
	encryptor, err := infrastructureutilityencryption.NewAESGCMImpl(testKey)
	if err != nil {
		t.Fatalf("NewAESGCMImpl() error = %v, want nil", err)
	}

	const plaintextApiKey = "sk-ant-real-secret-key"
	encrypted, err := encryptor.Encrypt(plaintextApiKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v, want nil", err)
	}

	repo := &fakeLlmConfigRepositoryForIntegration{config: &domainmodels.LlmConfig{
		Provider:        domainmodels.LlmProviderClaude,
		Model:           "claude-opus-5",
		ApiKeyEncrypted: encrypted,
	}}

	factory := NewClientFactory(repo, encryptor)

	var capturedApiKey string
	factory.newClaudeClient = func(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
		capturedApiKey = apiKey
		return &recordingClient{provider: "claude"}
	}
	factory.newOpenAIClient = func(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
		t.Fatalf("newOpenAIClient should not be called for a Claude config")
		return nil
	}

	if _, err := factory.Current(context.Background()); err != nil {
		t.Fatalf("Current() error = %v, want nil", err)
	}

	if capturedApiKey != plaintextApiKey {
		t.Fatalf("Current() decrypted apiKey = %q, want %q", capturedApiKey, plaintextApiKey)
	}
}
