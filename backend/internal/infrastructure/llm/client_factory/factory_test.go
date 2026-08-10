package infrastructurellmclientfactory

import (
	"context"
	"errors"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type stubLlmConfigRepository struct {
	config *domainmodels.LlmConfig
	err    error
}

func (s *stubLlmConfigRepository) Get(_ context.Context) (*domainmodels.LlmConfig, error) {
	return s.config, s.err
}
func (s *stubLlmConfigRepository) Upsert(_ context.Context, _ domainmodels.LlmProvider, _ string, _ []byte, _ *string, _ *uuid.UUID) (uuid.UUID, error) {
	panic("not used by this test")
}

type stubEncryptor struct{}

func (stubEncryptor) Encrypt(plaintext string) ([]byte, error)  { return []byte(plaintext), nil }
func (stubEncryptor) Decrypt(ciphertext []byte) (string, error) { return string(ciphertext), nil }

type recordingClient struct{ provider string }

func (r *recordingClient) GenerateText(_ context.Context, _ domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	return domaincontractsllm.GenerateTextResult{Text: r.provider}, nil
}

func TestCurrentBuildsClaudeAdapterForClaudeProvider(t *testing.T) {
	repo := &stubLlmConfigRepository{config: &domainmodels.LlmConfig{
		Provider:        domainmodels.LlmProviderClaude,
		Model:           "claude-opus-5",
		ApiKeyEncrypted: []byte("secret"),
	}}
	factory := &ClientFactory{repository: repo, encryptor: stubEncryptor{}}
	factory.newClaudeClient = func(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
		return &recordingClient{provider: "claude:" + apiKey + ":" + model}
	}
	factory.newOpenAIClient = func(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
		t.Fatalf("newOpenAIClient should not be called for a Claude config")
		return nil
	}

	client, err := factory.Current(context.Background())
	if err != nil {
		t.Fatalf("Current() error = %v, want nil", err)
	}

	result, err := client.GenerateText(context.Background(), domaincontractsllm.GenerateTextRequest{})
	if err != nil {
		t.Fatalf("GenerateText() error = %v, want nil", err)
	}
	if result.Text != "claude:secret:claude-opus-5" {
		t.Fatalf("GenerateText() Text = %q, want the decrypted key and model to have been passed through", result.Text)
	}
}

func TestCurrentBuildsOpenAIAdapterForOpenAIProvider(t *testing.T) {
	repo := &stubLlmConfigRepository{config: &domainmodels.LlmConfig{
		Provider:        domainmodels.LlmProviderOpenAI,
		Model:           "gpt-test",
		ApiKeyEncrypted: []byte("secret"),
	}}
	factory := &ClientFactory{repository: repo, encryptor: stubEncryptor{}}
	factory.newClaudeClient = func(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
		t.Fatalf("newClaudeClient should not be called for an OpenAI config")
		return nil
	}
	factory.newOpenAIClient = func(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
		return &recordingClient{provider: "openai:" + apiKey + ":" + model}
	}

	client, err := factory.Current(context.Background())
	if err != nil {
		t.Fatalf("Current() error = %v, want nil", err)
	}

	result, err := client.GenerateText(context.Background(), domaincontractsllm.GenerateTextRequest{})
	if err != nil {
		t.Fatalf("GenerateText() error = %v, want nil", err)
	}
	if result.Text != "openai:secret:gpt-test" {
		t.Fatalf("GenerateText() Text = %q, want the decrypted key and model to have been passed through", result.Text)
	}
}

func TestCurrentReturnsNotConfiguredWhenNoRowExists(t *testing.T) {
	repo := &stubLlmConfigRepository{config: nil}
	factory := NewClientFactory(repo, stubEncryptor{})

	_, err := factory.Current(context.Background())
	if !errors.Is(err, domainmodels.ErrTypeLlmConfigNotConfigured) {
		t.Fatalf("Current() error = %v, want ErrTypeLlmConfigNotConfigured", err)
	}
}
