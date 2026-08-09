package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type LlmProvider string

const (
	LlmProviderClaude LlmProvider = "CLAUDE"
	LlmProviderOpenAI LlmProvider = "OPENAI"
)

type LlmConfig struct {
	Id              uuid.UUID
	Provider        LlmProvider
	Model           string
	ApiKeyEncrypted []byte
	BaseURL         *string
	UpdatedAt       time.Time
}
