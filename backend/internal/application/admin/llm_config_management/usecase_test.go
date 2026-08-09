package applicationadminllmconfigmanagement

import (
	"context"
	"errors"
	"testing"

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

func TestUpdateRejectsUnknownProvider(t *testing.T) {
	repository := &recordingLlmConfigRepository{}
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &noopLogger{})

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
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &noopLogger{})

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
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &noopLogger{})

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
	usecase := NewUsecaseImpl(repository, recordingEncryptor{}, &noopLogger{})

	got, err := usecase.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("Get() = %+v, want nil when the repository has no config yet", got)
	}
}
