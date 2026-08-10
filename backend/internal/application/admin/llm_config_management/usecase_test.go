package applicationadminllmconfigmanagement

import (
	"context"
	"errors"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
	"github.com/google/uuid"
)

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
	generateErr error
}

func (f *fakeClient) GenerateText(_ context.Context, _ domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	if f.generateErr != nil {
		return domaincontractsllm.GenerateTextResult{}, f.generateErr
	}
	return domaincontractsllm.GenerateTextResult{Text: "OK"}, nil
}

type fakeClientFactory struct {
	client     *fakeClient
	currentErr error
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

func TestUpdateRejectsUnknownProvider(t *testing.T) {
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &fakeClientFactory{}, &noopLogger{})

	err := usecase.Update(context.Background(), domainusecasesadmin.UpdateLlmConfigRequest{
		Provider: domainmodels.LlmProvider("GEMINI"),
		Model:    "some-model",
		ApiKey:   "some-key",
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
		ApiKey:   "sk-ant-real-key",
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
