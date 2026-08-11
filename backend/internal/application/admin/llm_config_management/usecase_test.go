package applicationadminllmconfigmanagement

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
	"github.com/google/uuid"
)

func strPtr(value string) *string { return &value }

type recordingLlmConfigRepository struct {
	upsertCalls    int
	upsertedApiKey []byte
	upsertedModel  string
	upsertedProv   domainmodels.LlmProvider
	getConfig      *domainmodels.LlmConfig
}

func (r *recordingLlmConfigRepository) Get(_ context.Context) (*domainmodels.LlmConfig, error) {
	return r.getConfig, nil
}
func (r *recordingLlmConfigRepository) Upsert(_ context.Context, provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, _ *string, _ *uuid.UUID) (uuid.UUID, error) {
	r.upsertCalls++
	r.upsertedProv = provider
	r.upsertedModel = model
	r.upsertedApiKey = apiKeyEncrypted
	return uuid.New(), nil
}

type recordingEncryptor struct{}

func (recordingEncryptor) Encrypt(plaintext string) ([]byte, error) {
	return []byte("encrypted:" + plaintext), nil
}
func (recordingEncryptor) Decrypt(ciphertext []byte) (string, error) { return string(ciphertext), nil }

type noopLogger struct{ domaincontractslogger.Leveled }

func (f *noopLogger) Error(_ context.Context, _ string, _ string, _ domainmodels.LoggerMeta) {}
func (f *noopLogger) Warn(_ context.Context, _ string, _ string, _ domainmodels.LoggerMeta)  {}
func (f *noopLogger) Info(_ context.Context, _ string, _ string, _ domainmodels.LoggerMeta)  {}
func (f *noopLogger) Debug(_ context.Context, _ string, _ string, _ domainmodels.LoggerMeta) {}

type fakeClient struct {
	generateErr  error
	responseText string
	lastRequest  domaincontractsllm.GenerateTextRequest
}

func (f *fakeClient) GenerateText(_ context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	f.lastRequest = req
	if f.generateErr != nil {
		return domaincontractsllm.GenerateTextResult{}, f.generateErr
	}
	if f.responseText != "" {
		return domaincontractsllm.GenerateTextResult{Text: f.responseText}, nil
	}
	return domaincontractsllm.GenerateTextResult{Text: `{"status": "OK"}`}, nil
}

type fakeClientFactory struct {
	client              *fakeClient
	currentErr          error
	fromCredentialsErr  error
	fromCredentialsCall int
}

func (f *fakeClientFactory) Current(_ context.Context) (domaincontractsllm.Client, error) {
	if f.currentErr != nil {
		return nil, f.currentErr
	}
	if f.client != nil {
		return f.client, nil
	}
	return &fakeClient{}, nil
}

func (f *fakeClientFactory) FromCredentials(_ domainmodels.LlmProvider, _ string, _ *string, _ string) (domaincontractsllm.Client, error) {
	f.fromCredentialsCall++
	if f.fromCredentialsErr != nil {
		return nil, f.fromCredentialsErr
	}
	if f.client != nil {
		return f.client, nil
	}
	return &fakeClient{}, nil
}

func TestUpdateRejectsUnknownProvider(t *testing.T) {
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &fakeClientFactory{}, &noopLogger{})

	err := usecase.Update(context.Background(), domainusecasesadmin.UpdateLlmConfigRequest{
		Provider: domainmodels.LlmProvider("GROK"),
		Model:    "some-model",
		ApiKey:   strPtr("some-key"),
	})

	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("Update() error = %v, want validation error", err)
	}
	if repository.upsertCalls != 0 {
		t.Fatalf("repository Upsert() calls = %d, want 0", repository.upsertCalls)
	}
}

func TestUpdateEncryptsApiKeyBeforeStoring(t *testing.T) {
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &fakeClientFactory{}, &noopLogger{})

	err := usecase.Update(context.Background(), domainusecasesadmin.UpdateLlmConfigRequest{
		Provider: domainmodels.LlmProviderClaude,
		Model:    "claude-opus-5",
		ApiKey:   strPtr("sk-ant-real-key"),
	})

	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repository.upsertCalls != 1 {
		t.Fatalf("repository Upsert() calls = %d, want 1", repository.upsertCalls)
	}
	if string(repository.upsertedApiKey) != "encrypted:sk-ant-real-key" {
		t.Fatalf("repository Upsert() apiKeyEncrypted = %q, want the encrypted form, not the plaintext key", repository.upsertedApiKey)
	}
	if repository.upsertedProv != domainmodels.LlmProviderClaude || repository.upsertedModel != "claude-opus-5" {
		t.Fatalf("repository Upsert() provider/model = %q/%q, want CLAUDE/claude-opus-5", repository.upsertedProv, repository.upsertedModel)
	}
}

func TestUpdateKeepsExistingApiKeyWhenOmitted(t *testing.T) {
	existing := &domainmodels.LlmConfig{
		Provider:        domainmodels.LlmProviderClaude,
		Model:           "claude-sonnet-5",
		ApiKeyEncrypted: []byte("encrypted:previous-key"),
	}
	repository := &recordingLlmConfigRepository{getConfig: existing}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &fakeClientFactory{}, &noopLogger{})

	err := usecase.Update(context.Background(), domainusecasesadmin.UpdateLlmConfigRequest{
		Provider: domainmodels.LlmProviderClaude,
		Model:    "claude-opus-5",
		ApiKey:   nil,
	})

	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repository.upsertCalls != 1 {
		t.Fatalf("repository Upsert() calls = %d, want 1", repository.upsertCalls)
	}
	if string(repository.upsertedApiKey) != "encrypted:previous-key" {
		t.Fatalf("repository Upsert() apiKeyEncrypted = %q, want the previously stored key preserved", repository.upsertedApiKey)
	}
}

func TestUpdateRejectsOmittedApiKeyOnFirstTimeSetup(t *testing.T) {
	repository := &recordingLlmConfigRepository{getConfig: nil}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &fakeClientFactory{}, &noopLogger{})

	err := usecase.Update(context.Background(), domainusecasesadmin.UpdateLlmConfigRequest{
		Provider: domainmodels.LlmProviderClaude,
		Model:    "claude-opus-5",
		ApiKey:   nil,
	})

	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("Update() error = %v, want validation error", err)
	}
	if repository.upsertCalls != 0 {
		t.Fatalf("repository Upsert() calls = %d, want 0", repository.upsertCalls)
	}
}

func TestGetReturnsConfigUnchanged(t *testing.T) {
	config := &domainmodels.LlmConfig{Provider: domainmodels.LlmProviderOpenAI, Model: "gpt-test"}
	repository := &recordingLlmConfigRepository{getConfig: config}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &fakeClientFactory{}, &noopLogger{})

	got, err := usecase.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.Provider != domainmodels.LlmProviderOpenAI || got.Model != "gpt-test" {
		t.Fatalf("Get() = %+v, want the repository's config back unchanged", got)
	}
}

func TestGetReturnsNilWhenNoConfigExists(t *testing.T) {
	repository := &recordingLlmConfigRepository{getConfig: nil}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &fakeClientFactory{}, &noopLogger{})

	got, err := usecase.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("Get() = %+v, want nil when the repository has no config yet", got)
	}
}

func TestTestConnectionReturnsConnectedStatusOnSuccess(t *testing.T) {
	client := &fakeClient{}
	factory := &fakeClientFactory{client: client}
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	status, err := usecase.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("TestConnection() error = %v, want nil", err)
	}
	if status != domainmodels.LlmClientStatusConnected {
		t.Fatalf("TestConnection() status = %v, want CONNECTED", status)
	}
}

// Real generation calls always request structured JSON output (an
// object-rooted response_format schema), and a provider/model/base_url
// combination can support plain chat while rejecting that - confirmed
// live against OpenRouter, where a working chat completion still failed
// structured output with "schema must be type object". A ping that never
// exercises this exact code path is a false positive: it reports
// "Connected" for a configuration that generation will fail against.
func TestTestConnectionRequestsAnObjectRootedSchema(t *testing.T) {
	client := &fakeClient{}
	factory := &fakeClientFactory{client: client}
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	if _, err := usecase.TestConnection(context.Background()); err != nil {
		t.Fatalf("TestConnection() error = %v, want nil", err)
	}

	if client.lastRequest.ResponseSchema == nil {
		t.Fatal("GenerateText() request had no ResponseSchema — the ping doesn't exercise structured output, so it can't catch a config that fails on it")
	}
	var schema map[string]any
	if err := json.Unmarshal(client.lastRequest.ResponseSchema, &schema); err != nil {
		t.Fatalf("ResponseSchema is not valid JSON: %v", err)
	}
	if schema["type"] != "object" {
		t.Fatalf("ResponseSchema root type = %v, want \"object\" (OpenAI's Structured Outputs rejects any other root type)", schema["type"])
	}
}

// A provider that accepts the schema but doesn't actually honor it (garbled
// or missing output) should still fail the check, not just "the call didn't
// error" — ping() parses the response and confirms it round-tripped.
func TestTestConnectionFailsWhenStructuredResponseIsMalformed(t *testing.T) {
	client := &fakeClient{responseText: "not json"}
	factory := &fakeClientFactory{client: client}
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	_, err := usecase.TestConnection(context.Background())
	if err == nil {
		t.Fatal("TestConnection() error = nil, want an error when the structured response doesn't parse")
	}
}

func TestTestConnectionPropagatesResolveError(t *testing.T) {
	factory := &fakeClientFactory{currentErr: errors.New("no config configured")}
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	_, err := usecase.TestConnection(context.Background())
	if err == nil {
		t.Fatal("TestConnection() error = nil, want the factory's resolve error propagated")
	}
}

func TestTestConnectionPropagatesGenerateTextError(t *testing.T) {
	client := &fakeClient{generateErr: errors.New("invalid api key")}
	factory := &fakeClientFactory{client: client}
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	_, err := usecase.TestConnection(context.Background())
	if err == nil {
		t.Fatal("TestConnection() error = nil, want the client's GenerateText error propagated")
	}
}

func TestTestConnectionWithConfigUsesGivenApiKey(t *testing.T) {
	factory := &fakeClientFactory{client: &fakeClient{}}
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	status, err := usecase.TestConnectionWithConfig(context.Background(), domainusecasesadmin.TestLlmConnectionWithConfigRequest{
		Provider: domainmodels.LlmProviderOpenAI,
		Model:    "gpt-5",
		ApiKey:   strPtr("sk-given-key"),
	})

	if err != nil {
		t.Fatalf("TestConnectionWithConfig() error = %v, want nil", err)
	}
	if status != domainmodels.LlmClientStatusConnected {
		t.Fatalf("TestConnectionWithConfig() status = %v, want CONNECTED", status)
	}
	if factory.fromCredentialsCall != 1 {
		t.Fatalf("FromCredentials() calls = %d, want 1", factory.fromCredentialsCall)
	}
}

func TestTestConnectionWithConfigFallsBackToStoredKeyWhenOmitted(t *testing.T) {
	existing := &domainmodels.LlmConfig{
		Provider:        domainmodels.LlmProviderClaude,
		Model:           "claude-sonnet-5",
		ApiKeyEncrypted: []byte("encrypted:stored-key"),
	}
	factory := &fakeClientFactory{client: &fakeClient{}}
	repository := &recordingLlmConfigRepository{getConfig: existing}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	status, err := usecase.TestConnectionWithConfig(context.Background(), domainusecasesadmin.TestLlmConnectionWithConfigRequest{
		Provider: domainmodels.LlmProviderClaude,
		Model:    "claude-opus-5",
		ApiKey:   nil,
	})

	if err != nil {
		t.Fatalf("TestConnectionWithConfig() error = %v, want nil", err)
	}
	if status != domainmodels.LlmClientStatusConnected {
		t.Fatalf("TestConnectionWithConfig() status = %v, want CONNECTED", status)
	}
}

func TestTestConnectionWithConfigRejectsOmittedApiKeyWithNoExistingConfig(t *testing.T) {
	factory := &fakeClientFactory{client: &fakeClient{}}
	repository := &recordingLlmConfigRepository{getConfig: nil}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	_, err := usecase.TestConnectionWithConfig(context.Background(), domainusecasesadmin.TestLlmConnectionWithConfigRequest{
		Provider: domainmodels.LlmProviderClaude,
		Model:    "claude-opus-5",
		ApiKey:   nil,
	})

	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("TestConnectionWithConfig() error = %v, want validation error", err)
	}
	if factory.fromCredentialsCall != 0 {
		t.Fatalf("FromCredentials() calls = %d, want 0 (should fail before building a client)", factory.fromCredentialsCall)
	}
}

func TestTestConnectionWithConfigRejectsUnknownProvider(t *testing.T) {
	factory := &fakeClientFactory{client: &fakeClient{}}
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	_, err := usecase.TestConnectionWithConfig(context.Background(), domainusecasesadmin.TestLlmConnectionWithConfigRequest{
		Provider: domainmodels.LlmProvider("GROK"),
		Model:    "some-model",
		ApiKey:   strPtr("some-key"),
	})

	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("TestConnectionWithConfig() error = %v, want validation error", err)
	}
}

func TestTestConnectionWithConfigPropagatesGenerateTextError(t *testing.T) {
	factory := &fakeClientFactory{client: &fakeClient{generateErr: errors.New("model not supported by this base URL")}}
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, factory, &noopLogger{})

	_, err := usecase.TestConnectionWithConfig(context.Background(), domainusecasesadmin.TestLlmConnectionWithConfigRequest{
		Provider: domainmodels.LlmProviderGemini,
		Model:    "gemini-2.5-pro",
		ApiKey:   strPtr("some-key"),
	})

	if err == nil {
		t.Fatal("TestConnectionWithConfig() error = nil, want the client's GenerateText error propagated")
	}
}
