# LLM Provider Abstraction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give the backend a provider-agnostic way to make LLM calls (structured JSON or free-form text) where the active provider (Claude or OpenAI) and its credential are stored in Postgres, admin-configurable at runtime, and resolved fresh on every call — never a client built once at process startup.

**Architecture:** A dependency-free domain contract (`llm.Client`) with one method, two Postgres-backed adapters (Claude via `anthropic-sdk-go`, OpenAI via `openai-go`) implementing it, a `ClientFactory` that reads the single active `LlmConfig` row, decrypts its API key, and constructs the matching adapter per call, and a standard repository/usecase/handler stack (following this codebase's existing admin-feature conventions) to let an admin read (key redacted) and write that config.

**Tech Stack:** Go 1.26, pgx/v5 + squirrel (existing), golang-migrate SQL migrations, `crypto/aes`/`crypto/cipher` (stdlib, AES-GCM) for the API key at rest, `github.com/anthropics/anthropic-sdk-go` (new), `github.com/openai/openai-go` (new), Echo v5 (existing).

## Global Constraints

- Domain models and contracts stay dependency-free — no SDK imports outside `internal/infrastructure/`.
- Package names follow the existing flattened convention (e.g. `domainmodels`, `domaincontractsrepository`, `infrastructurerepositoryllmconfig`) — no nested Go packages with slashes in the name.
- Migrations are golang-migrate SQL files in `backend/database/migrations/`, named `YYYYMMDDHHMMSS_description.up.sql` / `.down.sql`.
- Validation is split: `internal/presentation/http/utils` validates raw shape (UUID parse, required-field presence); `internal/application/shared/validation.go` validates business rules (enum membership, string patterns, URL well-formedness).
- Every usecase method starts with `const tag = "<area>/<usecase>/<Method>"` and logs errors via the injected `domaincontractslogger.Leveled` with that tag.
- Test doubles are hand-written structs embedding the interface they fake (unimplemented methods panic by virtue of the nil embedded interface), never a mocking library.
- Verification bar for every task: `go build ./... && go vet ./... && gofmt -l .` clean, plus `go test -count=1 ./...` for anything with dedicated tests.
- The API key is never returned by any read endpoint, in any form — not full, not redacted-suffix. Only provider, model, base URL, and whether a key is currently set are ever read back.

---

## File structure

```
backend/
  database/migrations/
    20260809100000_llm_config.up.sql        (new)
    20260809100000_llm_config.down.sql      (new)
  internal/
    domain/
      models/
        llm_config.go                        (new — LlmProvider enum, LlmConfig struct)
        error.go                              (modify — add ErrTypeLlmConfigNotConfigured)
      contracts/
        repository/
          llm_config.go                       (new — LlmConfig repository interface)
        utility/
          encryption.go                       (new — Encryptor interface)
        llm/
          client.go                           (new — llm.Client interface, request/result types)
      usecases/
        admin/
          llm_config_management.go            (new — usecase interface + request DTOs)
    application/
      shared/
        validation.go                         (modify — add LLM-specific validators)
      admin/
        llm_config_management/
          usecase.go                          (new)
          usecase_test.go                     (new)
    infrastructure/
      repository/
        llm_config/
          postgres.go                         (new)
          postgres_query.go                   (new)
      utility/
        encryption/
          aesgcm.go                            (new)
          aesgcm_test.go                       (new)
      llm/
        claude/
          client.go                            (new)
          client_test.go                       (new)
        openai/
          client.go                            (new)
          client_test.go                       (new)
        factory.go                             (new)
        factory_test.go                        (new)
    config/
      env.go                                   (modify — add BE_LLM_ENCRYPTION_KEY)
    presentation/
      http/
        request/
          admin.go                             (modify — add LlmConfigPutRequest)
        response/
          llm_config.go                        (new)
        handler/
          admin/
            handler.go                         (modify — add LlmConfigGet/LlmConfigPut, new field)
        route/
          route.go                             (modify — add two routes + interface methods)
    composition/
      main/
        infrastructure.go                      (modify — wire repository, encryptor, factory)
        application.go                         (modify — wire usecase)
        presentation.go                        (modify — pass usecase to admin handler)
  database/seeder/
    permissions.json                           (modify — add llm_config:get / llm_config:set)
```

---

### Task 1: Migration and domain model for `LlmConfig`

**Files:**
- Create: `backend/database/migrations/20260809100000_llm_config.up.sql`
- Create: `backend/database/migrations/20260809100000_llm_config.down.sql`
- Create: `backend/internal/domain/models/llm_config.go`
- Modify: `backend/internal/domain/models/error.go`

**Interfaces:**
- Produces: `domainmodels.LlmProvider` (`"CLAUDE"` | `"OPENAI"`), `domainmodels.LlmConfig{Id, Provider, Model, ApiKeyEncrypted, BaseURL, UpdatedAt}`, `domainmodels.ErrTypeLlmConfigNotConfigured`.

- [ ] **Step 1: Write the migration**

`backend/database/migrations/20260809100000_llm_config.up.sql`:

```sql
CREATE TABLE llm_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    api_key_encrypted BYTEA NOT NULL,
    base_url TEXT,
    singleton BOOLEAN NOT NULL DEFAULT TRUE UNIQUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_llm_config_singleton CHECK (singleton = TRUE),
    CONSTRAINT chk_llm_config_provider CHECK (provider IN ('CLAUDE', 'OPENAI'))
);
```

The `singleton BOOLEAN UNIQUE` column (always `TRUE`, enforced by the check constraint) means the table can never hold more than one row — a plain `UNIQUE` constraint gives us `ON CONFLICT` upsert semantics without hand-rolled delete-then-insert transaction logic.

`backend/database/migrations/20260809100000_llm_config.down.sql`:

```sql
DROP TABLE IF EXISTS llm_config;
```

- [ ] **Step 2: Apply the migration locally**

Run: `migrate -path backend/database/migrations -database "$DATABASE_URL" up`
Expected: no error; `\d llm_config` in `psql` shows the table above.

- [ ] **Step 3: Add the domain model**

`backend/internal/domain/models/llm_config.go`:

```go
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
```

- [ ] **Step 4: Add the not-configured error sentinel**

In `backend/internal/domain/models/error.go`, add alongside the existing `ErrType*` sentinels (same file, same `errors.New("SOME_CODE")` pattern used for `ErrTypeNotFound` etc.):

```go
var ErrTypeLlmConfigNotConfigured = errors.New("LLM_CONFIG_NOT_CONFIGURED")
```

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/models/`
Expected: no output from any command (clean build, no vet warnings, no formatting diffs).

- [ ] **Step 6: Commit**

```bash
git add backend/database/migrations/20260809100000_llm_config.up.sql \
        backend/database/migrations/20260809100000_llm_config.down.sql \
        backend/internal/domain/models/llm_config.go \
        backend/internal/domain/models/error.go
git commit -m "feat: add llm_config table and domain model"
```

---

### Task 2: `LlmConfig` repository (contract + Postgres implementation)

**Files:**
- Create: `backend/internal/domain/contracts/repository/llm_config.go`
- Create: `backend/internal/infrastructure/repository/llm_config/postgres.go`
- Create: `backend/internal/infrastructure/repository/llm_config/postgres_query.go`

**Interfaces:**
- Consumes: `domainmodels.LlmConfig`, `domainmodels.LlmProvider` (Task 1); `infrastructurerepositoryshared.BasePostgres`, `MapPgxError`, `NotFound` (existing).
- Produces: `domaincontractsrepository.LlmConfig` interface with `Get(ctx) (*domainmodels.LlmConfig, error)` and `Upsert(ctx, provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, baseURL *string) (id uuid.UUID, err error)`.

This layer has no dedicated unit test in this codebase's convention (repositories are exercised via the usecase-layer fakes in later tasks, not directly) — verification is `go build`/`go vet` only, same as every other `postgres.go` in the repo.

- [ ] **Step 1: Write the domain contract**

`backend/internal/domain/contracts/repository/llm_config.go`:

```go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type LlmConfig interface {
	Get(ctx context.Context) (*domainmodels.LlmConfig, error)
	Upsert(ctx context.Context, provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, baseURL *string) (id uuid.UUID, err error)
}
```

- [ ] **Step 2: Write the query builders**

`backend/internal/infrastructure/repository/llm_config/postgres_query.go`:

```go
package infrastructurerepositoryllmconfig

import (
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func (p *postgresImpl) queryGet() (query string, args []any, err error) {
	return p.SqrD.Select("id", "provider", "model", "api_key_encrypted", "base_url", "updated_at").
		From("llm_config").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryUpsert(provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, baseURL *string) (query string, args []any, err error) {
	return p.SqrD.Insert("llm_config").
		Columns("id", "provider", "model", "api_key_encrypted", "base_url", "singleton").
		Values(uuid.New(), string(provider), model, apiKeyEncrypted, baseURL, true).
		Suffix(`ON CONFLICT (singleton) DO UPDATE SET
			provider = EXCLUDED.provider,
			model = EXCLUDED.model,
			api_key_encrypted = EXCLUDED.api_key_encrypted,
			base_url = EXCLUDED.base_url,
			updated_at = NOW()
			RETURNING id`).
		ToSql()
}
```

- [ ] **Step 3: Write the Postgres implementation**

`backend/internal/infrastructure/repository/llm_config/postgres.go`:

```go
package infrastructurerepositoryllmconfig

import (
	"context"

	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(
	dt pgxdt.Pgxdt,
	sqrQuestion *squirrel.StatementBuilderType,
	sqrDollar *squirrel.StatementBuilderType,
) domaincontractsrepository.LlmConfig {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) Get(ctx context.Context) (*domainmodels.LlmConfig, error) {
	query, args, err := p.queryGet()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build llm_config get query", err)
	}

	row := p.Dt.QueryRow(ctx, query, args...)
	config, err := scanPgxLlmConfig(row)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read llm_config", err)
	}
	return &config, nil
}

func (p *postgresImpl) Upsert(ctx context.Context, provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, baseURL *string) (uuid.UUID, error) {
	query, args, err := p.queryUpsert(provider, model, apiKeyEncrypted, baseURL)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build llm_config upsert query", err)
	}

	row := p.Dt.QueryRow(ctx, query, args...)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to upsert llm_config", err)
	}
	return id, nil
}

func scanPgxLlmConfig(row interface {
	Scan(dest ...any) error
}) (domainmodels.LlmConfig, error) {
	var config domainmodels.LlmConfig
	var provider string
	if err := row.Scan(&config.Id, &provider, &config.Model, &config.ApiKeyEncrypted, &config.BaseURL, &config.UpdatedAt); err != nil {
		return domainmodels.LlmConfig{}, err
	}
	config.Provider = domainmodels.LlmProvider(provider)
	return config, nil
}
```

If `p.Dt.QueryRow` returns a concrete `pgx.Row` rather than the ad-hoc `Scan`-only interface used above, change `scanPgxLlmConfig`'s parameter type to `pgx.Row` and add the `github.com/jackc/pgx/v5` import — match whichever existing `scanPgx*` function in `internal/infrastructure/repository/` this repo's `pgxdt.Pgxdt.QueryRow` actually returns.

- [ ] **Step 4: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/llm_config/`
Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/contracts/repository/llm_config.go \
        backend/internal/infrastructure/repository/llm_config/
git commit -m "feat: add LlmConfig repository"
```

---

### Task 3: AES-GCM encryption utility

**Files:**
- Create: `backend/internal/domain/contracts/utility/encryption.go`
- Create: `backend/internal/infrastructure/utility/encryption/aesgcm.go`
- Create: `backend/internal/infrastructure/utility/encryption/aesgcm_test.go`
- Modify: `backend/internal/config/env.go`

**Interfaces:**
- Produces: `domaincontractsutility.Encryptor` interface with `Encrypt(plaintext string) (ciphertext []byte, err error)` / `Decrypt(ciphertext []byte) (plaintext string, err error)`; `config.LlmEncryptionKey string`.

- [ ] **Step 1: Write the failing test**

`backend/internal/infrastructure/utility/encryption/aesgcm_test.go`:

```go
package infrastructureutilityencryption

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := "01234567890123456789012345678901" // 32 bytes for AES-256
	encryptor, err := NewAESGCMImpl(key)
	if err != nil {
		t.Fatalf("NewAESGCMImpl() error = %v, want nil", err)
	}

	plaintext := "sk-ant-super-secret-key"
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v, want nil", err)
	}
	if bytes.Contains(ciphertext, []byte(plaintext)) {
		t.Fatalf("Encrypt() output contains the plaintext verbatim")
	}

	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v, want nil", err)
	}
	if decrypted != plaintext {
		t.Fatalf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptProducesDistinctCiphertextsForSamePlaintext(t *testing.T) {
	key := "01234567890123456789012345678901"
	encryptor, err := NewAESGCMImpl(key)
	if err != nil {
		t.Fatalf("NewAESGCMImpl() error = %v, want nil", err)
	}

	first, err := encryptor.Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt() error = %v, want nil", err)
	}
	second, err := encryptor.Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt() error = %v, want nil", err)
	}
	if bytes.Equal(first, second) {
		t.Fatalf("Encrypt() produced identical ciphertexts for the same plaintext across two calls (nonce reuse)")
	}
}

func TestDecryptRejectsCorruptedCiphertext(t *testing.T) {
	key := "01234567890123456789012345678901"
	encryptor, err := NewAESGCMImpl(key)
	if err != nil {
		t.Fatalf("NewAESGCMImpl() error = %v, want nil", err)
	}

	ciphertext, err := encryptor.Encrypt("plaintext")
	if err != nil {
		t.Fatalf("Encrypt() error = %v, want nil", err)
	}
	ciphertext[len(ciphertext)-1] ^= 0xFF // flip the last byte

	if _, err := encryptor.Decrypt(ciphertext); err == nil {
		t.Fatalf("Decrypt() error = nil, want an authentication error on corrupted ciphertext")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test ./internal/infrastructure/utility/encryption/... -v`
Expected: FAIL — `NewAESGCMImpl` undefined.

- [ ] **Step 3: Write the domain contract**

`backend/internal/domain/contracts/utility/encryption.go`:

```go
package domaincontractsutility

type Encryptor interface {
	Encrypt(plaintext string) (ciphertext []byte, err error)
	Decrypt(ciphertext []byte) (plaintext string, err error)
}
```

- [ ] **Step 4: Write the AES-GCM implementation**

`backend/internal/infrastructure/utility/encryption/aesgcm.go`:

```go
package infrastructureutilityencryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
)

type aesGCMImpl struct {
	gcm cipher.AEAD
}

// NewAESGCMImpl builds an AES-256-GCM encryptor. key must be exactly 32 bytes
// (AES-256); pass it via the BE_LLM_ENCRYPTION_KEY environment variable.
func NewAESGCMImpl(key string) (domaincontractsutility.Encryptor, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("failed to construct AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to construct GCM mode: %w", err)
	}
	return &aesGCMImpl{gcm: gcm}, nil
}

func (a *aesGCMImpl) Encrypt(plaintext string) ([]byte, error) {
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return a.gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func (a *aesGCMImpl) Decrypt(ciphertext []byte) (string, error) {
	nonceSize := a.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext shorter than nonce size")
	}
	nonce, sealed := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := a.gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plaintext), nil
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `cd backend && go test ./internal/infrastructure/utility/encryption/... -v`
Expected: PASS on all three tests.

- [ ] **Step 6: Add the encryption key to env config**

In `backend/internal/config/env.go`, add alongside the existing `var` block and `LoadEnv()` overrides (same pattern as `PostgresHost`):

```go
var LlmEncryptionKey string = ""
```

```go
LlmEncryptionKey = envGetString("BE_LLM_ENCRYPTION_KEY", LlmEncryptionKey)
```

Add `BE_LLM_ENCRYPTION_KEY=CHANGE_ME_32_BYTE_KEY_1234567890` to `.env.example`, in the same section style as the other `BE_*` groups, with a comment noting it must be exactly 32 bytes for AES-256.

- [ ] **Step 7: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/utility/ internal/infrastructure/utility/encryption/ internal/config/`
Expected: no output.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/domain/contracts/utility/encryption.go \
        backend/internal/infrastructure/utility/encryption/ \
        backend/internal/config/env.go \
        backend/.env.example
git commit -m "feat: add AES-GCM encryption utility for secrets at rest"
```

---

### Task 4: Validation helpers for LLM config fields

**Files:**
- Modify: `backend/internal/application/shared/validation.go`
- Create: `backend/internal/application/shared/validation_llm_config_test.go`

**Interfaces:**
- Consumes: `domainmodels.LlmProvider` (Task 1).
- Produces: `applicationshared.RequiredLlmProvider(value domainmodels.LlmProvider, field string) (domainmodels.LlmProvider, error)`, `applicationshared.RequiredLlmModel(value string, field string) (string, error)`, `applicationshared.OptionalLlmBaseURL(value *string, field string) (*string, error)` (delegates to the existing `RequiredURL`), `applicationshared.RequiredLlmApiKey(value string, field string) (string, error)`.

- [ ] **Step 1: Write the failing tests**

`backend/internal/application/shared/validation_llm_config_test.go`:

```go
package applicationshared

import (
	"errors"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

func TestRequiredLlmProviderRejectsUnknownValue(t *testing.T) {
	_, err := RequiredLlmProvider(domainmodels.LlmProvider("GEMINI"), "provider")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredLlmProvider() error = %v, want validation error", err)
	}
}

func TestRequiredLlmProviderAcceptsKnownValues(t *testing.T) {
	for _, provider := range []domainmodels.LlmProvider{domainmodels.LlmProviderClaude, domainmodels.LlmProviderOpenAI} {
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
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/shared/... -run TestRequiredLlm -v`
Expected: FAIL — `RequiredLlmProvider` (etc.) undefined.

- [ ] **Step 3: Write the validators**

Add to `backend/internal/application/shared/validation.go`, following the existing `validPayloadSchemaDefinitionTypes`-style membership-map pattern:

```go
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
```

If `fmt` is not already imported in `validation.go`, add it — the existing validators in this file already build error strings this way, so it likely already is.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/shared/... -run TestRequiredLlm -v && go test ./internal/application/shared/... -run TestOptionalLlm -v`
Expected: PASS on all six tests.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/shared/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/shared/validation.go \
        backend/internal/application/shared/validation_llm_config_test.go
git commit -m "feat: add validators for LLM config fields"
```

---

### Task 5: `llm.Client` domain contract

**Files:**
- Create: `backend/internal/domain/contracts/llm/client.go`

**Interfaces:**
- Produces: `llm.Client` interface, `llm.GenerateTextRequest{System, Prompt, MaxOutputTokens, ResponseSchema}`, `llm.GenerateTextResult{Text, Usage}`, `llm.Usage{InputTokens, OutputTokens int32}`.

This is a pure interface/type declaration with nothing to unit-test on its own — Tasks 6 and 7 exercise it directly against real adapters.

- [ ] **Step 1: Write the contract**

`backend/internal/domain/contracts/llm/client.go`:

```go
package domaincontractsllm

import (
	"context"
	"encoding/json"
)

// Client is a provider-agnostic interface for making a single LLM generation
// call. Implementations live in internal/infrastructure/llm/<provider>/ and
// are constructed per-call by the ClientFactory (internal/infrastructure/llm),
// never held as a process-lifetime singleton.
type Client interface {
	// GenerateText makes one generation call. When req.ResponseSchema is nil,
	// the result is free-form text (e.g. generated code). When it is set, the
	// result's Text field is a JSON string conforming to that schema.
	GenerateText(ctx context.Context, req GenerateTextRequest) (GenerateTextResult, error)
}

type GenerateTextRequest struct {
	System          string
	Prompt          string
	MaxOutputTokens int32
	ResponseSchema  json.RawMessage
}

type GenerateTextResult struct {
	Text  string
	Usage Usage
}

type Usage struct {
	InputTokens  int32
	OutputTokens int32
}
```

- [ ] **Step 2: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/llm/`
Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/domain/contracts/llm/client.go
git commit -m "feat: add provider-agnostic llm.Client contract"
```

---

### Task 6: Claude adapter

**Files:**
- Create: `backend/internal/infrastructure/llm/claude/client.go`
- Create: `backend/internal/infrastructure/llm/claude/client_test.go`
- Modify: `backend/go.mod` (add `github.com/anthropics/anthropic-sdk-go`)

**Interfaces:**
- Consumes: `domaincontractsllm.Client`, `GenerateTextRequest`, `GenerateTextResult`, `Usage` (Task 5).
- Produces: `infrastructurellmclaude.NewClient(apiKey string, baseURL *string, model string) domaincontractsllm.Client`.

- [ ] **Step 1: Add the dependency**

Run: `cd backend && go get github.com/anthropics/anthropic-sdk-go@latest`

- [ ] **Step 2: Write the failing test**

`backend/internal/infrastructure/llm/claude/client_test.go` spins up a fake Anthropic API server so the test never makes a real network call. It asserts both the free-form path and the structured-output path.

```go
package infrastructurellmclaude

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
)

func TestGenerateTextFreeForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "msg_test",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "text", "text": "function encode(state) { return []; }"}],
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 120, "output_tokens": 45}
		}`))
	}))
	defer server.Close()

	baseURL := server.URL
	client := NewClient("test-api-key", &baseURL, "claude-opus-5")

	result, err := client.GenerateText(context.Background(), domaincontractsllm.GenerateTextRequest{
		System:          "You write JavaScript IR encoders.",
		Prompt:          "Write an encoder for POWER only.",
		MaxOutputTokens: 4096,
	})
	if err != nil {
		t.Fatalf("GenerateText() error = %v, want nil", err)
	}
	if result.Text != "function encode(state) { return []; }" {
		t.Fatalf("GenerateText() Text = %q, want the generated function body", result.Text)
	}
	if result.Usage.InputTokens != 120 || result.Usage.OutputTokens != 45 {
		t.Fatalf("GenerateText() Usage = %+v, want {120 45}", result.Usage)
	}
}

func TestGenerateTextStructuredSendsOutputConfigFormat(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "msg_test",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"description\":\"set to cool\"}"}],
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 80, "output_tokens": 20}
		}`))
	}))
	defer server.Close()

	baseURL := server.URL
	client := NewClient("test-api-key", &baseURL, "claude-opus-5")

	schema := json.RawMessage(`{"type":"object","properties":{"description":{"type":"string"}},"required":["description"],"additionalProperties":false}`)
	result, err := client.GenerateText(context.Background(), domaincontractsllm.GenerateTextRequest{
		Prompt:          "Describe the case.",
		MaxOutputTokens: 256,
		ResponseSchema:  schema,
	})
	if err != nil {
		t.Fatalf("GenerateText() error = %v, want nil", err)
	}
	if result.Text != `{"description":"set to cool"}` {
		t.Fatalf("GenerateText() Text = %q, want the raw JSON string", result.Text)
	}

	outputConfig, ok := capturedBody["output_config"].(map[string]any)
	if !ok {
		t.Fatalf("request body has no output_config field; body = %+v", capturedBody)
	}
	format, ok := outputConfig["format"].(map[string]any)
	if !ok || format["type"] != "json_schema" {
		t.Fatalf("output_config.format = %+v, want {type: json_schema, ...}", format)
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/infrastructure/llm/claude/... -v`
Expected: FAIL — `NewClient` undefined.

- [ ] **Step 4: Write the implementation**

`backend/internal/infrastructure/llm/claude/client.go`:

```go
package infrastructurellmclaude

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
)

type claudeClient struct {
	sdk   anthropic.Client
	model string
}

// NewClient constructs a Claude-backed llm.Client. Called per-call by the
// ClientFactory (internal/infrastructure/llm) — never held as a singleton,
// so a credential/provider change takes effect on the very next call.
func NewClient(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != nil {
		opts = append(opts, option.WithBaseURL(*baseURL))
	}
	return &claudeClient{
		sdk:   anthropic.NewClient(opts...),
		model: model,
	}
}

func (c *claudeClient) GenerateText(ctx context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: int64(req.MaxOutputTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	}
	if req.System != "" {
		params.System = []anthropic.TextBlockParam{{Text: req.System}}
	}
	if len(req.ResponseSchema) > 0 {
		params.OutputConfig = anthropic.OutputConfigParam{
			Format: anthropic.OutputFormatParamUnion{
				OfJSONOutputFormat: &anthropic.JSONOutputFormatParam{
					Schema: req.ResponseSchema,
				},
			},
		}
	}

	message, err := c.sdk.Messages.New(ctx, params)
	if err != nil {
		return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("claude generation failed: %w", err)
	}

	var text string
	for _, block := range message.Content {
		if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
			text = textBlock.Text
			break
		}
	}

	return domaincontractsllm.GenerateTextResult{
		Text: text,
		Usage: domaincontractsllm.Usage{
			InputTokens:  int32(message.Usage.InputTokens),
			OutputTokens: int32(message.Usage.OutputTokens),
		},
	}, nil
}
```

**If the compiler rejects `OutputConfigParam` / `OutputFormatParamUnion` / `JSONOutputFormatParam`:** these three type names are inferred from the naming pattern used consistently across every other Anthropic Go SDK param union (`Of<Variant>` fields on a `<Feature>ParamUnion`, e.g. `ThinkingConfigParamUnion{OfAdaptive: ...}`), not confirmed against the SDK source. Run `grep -r "OutputFormat\|OutputConfig" $(go env GOMODCACHE)/github.com/anthropics/anthropic-sdk-go@*/` to find the actual type names, fix the three references, and re-run the test — the request/response shape (an `output_config.format` object with `type: "json_schema"` and a `schema` field) is confirmed from the API docs and is what the test in Step 2 asserts against, so only the Go type names might need adjusting, not the logic.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/infrastructure/llm/claude/... -v`
Expected: PASS on both tests.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/infrastructure/llm/claude/`
Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add backend/go.mod backend/go.sum backend/internal/infrastructure/llm/claude/
git commit -m "feat: add Claude adapter for llm.Client"
```

---

### Task 7: OpenAI adapter

**Files:**
- Create: `backend/internal/infrastructure/llm/openai/client.go`
- Create: `backend/internal/infrastructure/llm/openai/client_test.go`
- Modify: `backend/go.mod` (add `github.com/openai/openai-go`)

**Interfaces:**
- Consumes: `domaincontractsllm.Client`, `GenerateTextRequest`, `GenerateTextResult`, `Usage` (Task 5).
- Produces: `infrastructurellmopenai.NewClient(apiKey string, baseURL *string, model string) domaincontractsllm.Client`.

- [ ] **Step 1: Add the dependency**

Run: `cd backend && go get github.com/openai/openai-go@latest`

- [ ] **Step 2: Write the failing test**

`backend/internal/infrastructure/llm/openai/client_test.go`, same fake-server structure as the Claude test:

```go
package infrastructurellmopenai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
)

func TestGenerateTextFreeForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "chatcmpl_test",
			"object": "chat.completion",
			"model": "gpt-test",
			"choices": [{"index": 0, "finish_reason": "stop", "message": {"role": "assistant", "content": "function encode(state) { return []; }"}}],
			"usage": {"prompt_tokens": 100, "completion_tokens": 40, "total_tokens": 140}
		}`))
	}))
	defer server.Close()

	baseURL := server.URL
	client := NewClient("test-api-key", &baseURL, "gpt-test")

	result, err := client.GenerateText(context.Background(), domaincontractsllm.GenerateTextRequest{
		System:          "You write JavaScript IR encoders.",
		Prompt:          "Write an encoder for POWER only.",
		MaxOutputTokens: 4096,
	})
	if err != nil {
		t.Fatalf("GenerateText() error = %v, want nil", err)
	}
	if result.Text != "function encode(state) { return []; }" {
		t.Fatalf("GenerateText() Text = %q, want the generated function body", result.Text)
	}
	if result.Usage.InputTokens != 100 || result.Usage.OutputTokens != 40 {
		t.Fatalf("GenerateText() Usage = %+v, want {100 40}", result.Usage)
	}
}

func TestGenerateTextStructuredSendsResponseFormat(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "chatcmpl_test",
			"object": "chat.completion",
			"model": "gpt-test",
			"choices": [{"index": 0, "finish_reason": "stop", "message": {"role": "assistant", "content": "{\"description\":\"set to cool\"}"}}],
			"usage": {"prompt_tokens": 60, "completion_tokens": 15, "total_tokens": 75}
		}`))
	}))
	defer server.Close()

	baseURL := server.URL
	client := NewClient("test-api-key", &baseURL, "gpt-test")

	schema := json.RawMessage(`{"type":"object","properties":{"description":{"type":"string"}},"required":["description"],"additionalProperties":false}`)
	result, err := client.GenerateText(context.Background(), domaincontractsllm.GenerateTextRequest{
		Prompt:          "Describe the case.",
		MaxOutputTokens: 256,
		ResponseSchema:  schema,
	})
	if err != nil {
		t.Fatalf("GenerateText() error = %v, want nil", err)
	}
	if result.Text != `{"description":"set to cool"}` {
		t.Fatalf("GenerateText() Text = %q, want the raw JSON string", result.Text)
	}

	responseFormat, ok := capturedBody["response_format"].(map[string]any)
	if !ok || responseFormat["type"] != "json_schema" {
		t.Fatalf("response_format = %+v, want {type: json_schema, ...}", responseFormat)
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/infrastructure/llm/openai/... -v`
Expected: FAIL — `NewClient` undefined.

- [ ] **Step 4: Write the implementation**

`backend/internal/infrastructure/llm/openai/client.go`:

```go
package infrastructurellmopenai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
)

type openAIClient struct {
	sdk   openai.Client
	model string
}

// NewClient constructs an OpenAI-backed llm.Client. Called per-call by the
// ClientFactory (internal/infrastructure/llm) — never held as a singleton.
func NewClient(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != nil {
		opts = append(opts, option.WithBaseURL(*baseURL))
	}
	return &openAIClient{
		sdk:   openai.NewClient(opts...),
		model: model,
	}
}

func (c *openAIClient) GenerateText(ctx context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	params := openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(req.Prompt),
		},
		MaxTokens: openai.Int(int64(req.MaxOutputTokens)),
	}
	if req.System != "" {
		params.Messages = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(req.System)}, params.Messages...)
	}
	if len(req.ResponseSchema) > 0 {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "response",
					Schema: req.ResponseSchema,
					Strict: openai.Bool(true),
				},
			},
		}
	}

	completion, err := c.sdk.Chat.Completions.New(ctx, params)
	if err != nil {
		return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("openai generation failed: %w", err)
	}

	var text string
	if len(completion.Choices) > 0 {
		text = completion.Choices[0].Message.Content
	}

	return domaincontractsllm.GenerateTextResult{
		Text: text,
		Usage: domaincontractsllm.Usage{
			InputTokens:  int32(completion.Usage.PromptTokens),
			OutputTokens: int32(completion.Usage.CompletionTokens),
		},
	}, nil
}
```

**If the compiler rejects any of `ChatCompletionNewParamsResponseFormatUnion` / `ResponseFormatJSONSchemaParam` / `ResponseFormatJSONSchemaJSONSchemaParam`:** same caveat as Task 6 — these names follow the SDK's consistent `Of<Variant>`-union codegen pattern but aren't confirmed against source. Run `grep -r "ResponseFormat" $(go env GOMODCACHE)/github.com/openai/openai-go@*/` to find the actual names and fix the references; the request shape (`response_format: {type: "json_schema", json_schema: {name, schema, strict}}`) is what Step 2's test asserts against and does not need to change.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/infrastructure/llm/openai/... -v`
Expected: PASS on both tests.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/infrastructure/llm/openai/`
Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add backend/go.mod backend/go.sum backend/internal/infrastructure/llm/openai/
git commit -m "feat: add OpenAI adapter for llm.Client"
```

---

### Task 8: `ClientFactory`

**Files:**
- Create: `backend/internal/infrastructure/llm/factory.go`
- Create: `backend/internal/infrastructure/llm/factory_test.go`

**Interfaces:**
- Consumes: `domaincontractsrepository.LlmConfig` (Task 2), `domaincontractsutility.Encryptor` (Task 3), `domaincontractsllm.Client` (Task 5), `infrastructurellmclaude.NewClient` (Task 6), `infrastructurellmopenai.NewClient` (Task 7).
- Produces: `infrastructurellm.NewClientFactory(repo, encryptor) *ClientFactory` and `(*ClientFactory).Current(ctx) (domaincontractsllm.Client, error)`.

The factory's real constructor functions (`infrastructurellmclaude.NewClient`, `infrastructurellmopenai.NewClient`) are injected as fields with sensible defaults, so the test below can substitute fakes without constructing a real SDK client or hitting the network.

- [ ] **Step 1: Write the failing test**

`backend/internal/infrastructure/llm/factory_test.go`:

```go
package infrastructurellm

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
func (s *stubLlmConfigRepository) Upsert(_ context.Context, _ domainmodels.LlmProvider, _ string, _ []byte, _ *string) (uuid.UUID, error) {
	panic("not used by this test")
}

type stubEncryptor struct{}

func (stubEncryptor) Encrypt(plaintext string) ([]byte, error) { return []byte(plaintext), nil }
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
	factory := NewClientFactory(repo, stubEncryptor{})
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
	factory := NewClientFactory(repo, stubEncryptor{})
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
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test ./internal/infrastructure/llm/... -run TestCurrent -v`
Expected: FAIL — `NewClientFactory` undefined.

- [ ] **Step 3: Write the implementation**

`backend/internal/infrastructure/llm/factory.go`:

```go
package infrastructurellm

import (
	"context"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurellmclaude "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/llm/claude"
	infrastructurellmopenai "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/llm/openai"
)

type newClientFunc func(apiKey string, baseURL *string, model string) domaincontractsllm.Client

// ClientFactory resolves the current LlmConfig and builds the matching
// adapter fresh on every call — it deliberately holds no client instance
// between calls, so a config change takes effect on the very next one.
type ClientFactory struct {
	repository domaincontractsrepository.LlmConfig
	encryptor  domaincontractsutility.Encryptor

	newClaudeClient newClientFunc
	newOpenAIClient newClientFunc
}

func NewClientFactory(repository domaincontractsrepository.LlmConfig, encryptor domaincontractsutility.Encryptor) *ClientFactory {
	return &ClientFactory{
		repository:      repository,
		encryptor:       encryptor,
		newClaudeClient: infrastructurellmclaude.NewClient,
		newOpenAIClient: infrastructurellmopenai.NewClient,
	}
}

func (f *ClientFactory) Current(ctx context.Context) (domaincontractsllm.Client, error) {
	config, err := f.repository.Get(ctx)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, domainmodels.ErrTypeLlmConfigNotConfigured
	}

	apiKey, err := f.encryptor.Decrypt(config.ApiKeyEncrypted)
	if err != nil {
		return nil, err
	}

	switch config.Provider {
	case domainmodels.LlmProviderClaude:
		return f.newClaudeClient(apiKey, config.BaseURL, config.Model), nil
	case domainmodels.LlmProviderOpenAI:
		return f.newOpenAIClient(apiKey, config.BaseURL, config.Model), nil
	default:
		return nil, domainmodels.NewError("llm_config has an unrecognized provider", domainmodels.ErrTypeFailure, nil)
	}
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/infrastructure/llm/... -v`
Expected: PASS on all three tests.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/infrastructure/llm/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/infrastructure/llm/factory.go backend/internal/infrastructure/llm/factory_test.go
git commit -m "feat: add per-call LLM ClientFactory"
```

---

### Task 9: `LlmConfigManagement` usecase

**Files:**
- Create: `backend/internal/domain/usecases/admin/llm_config_management.go`
- Create: `backend/internal/application/admin/llm_config_management/usecase.go`
- Create: `backend/internal/application/admin/llm_config_management/usecase_test.go`

**Interfaces:**
- Consumes: `domaincontractsrepository.LlmConfig` (Task 2), `domaincontractsutility.Encryptor` (Task 3), `applicationshared.RequiredLlmProvider`/`RequiredLlmModel`/`RequiredLlmApiKey`/`OptionalLlmBaseURL` (Task 4).
- Produces: `domainusecasesadmin.LlmConfigManagement` interface with `Get(ctx) (*domainmodels.LlmConfig, error)` and `Update(ctx, request UpdateLlmConfigRequest) error`; `domainusecasesadmin.UpdateLlmConfigRequest{Provider, Model, ApiKey, BaseURL}`.

- [ ] **Step 1: Write the usecase interface and request DTO**

`backend/internal/domain/usecases/admin/llm_config_management.go`:

```go
package domainusecasesadmin

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type LlmConfigManagement interface {
	Get(ctx context.Context) (*domainmodels.LlmConfig, error)
	Update(ctx context.Context, request UpdateLlmConfigRequest) error
}

type UpdateLlmConfigRequest struct {
	Provider domainmodels.LlmProvider
	Model    string
	ApiKey   string
	BaseURL  *string
}
```

- [ ] **Step 2: Write the failing tests**

`backend/internal/application/admin/llm_config_management/usecase_test.go`:

```go
package applicationadminllmconfigmanagement

import (
	"context"
	"errors"
	"testing"

	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type recordingLlmConfigRepository struct {
	upsertCalls     int
	upsertedApiKey  []byte
	upsertedModel   string
	upsertedProv    domainmodels.LlmProvider
	getConfig       *domainmodels.LlmConfig
}

func (r *recordingLlmConfigRepository) Get(_ context.Context) (*domainmodels.LlmConfig, error) {
	return r.getConfig, nil
}
func (r *recordingLlmConfigRepository) Upsert(_ context.Context, provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, _ *string) (uuid.UUID, error) {
	r.upsertCalls++
	r.upsertedProv = provider
	r.upsertedModel = model
	r.upsertedApiKey = apiKeyEncrypted
	return uuid.New(), nil
}

type recordingEncryptor struct{}

func (recordingEncryptor) Encrypt(plaintext string) ([]byte, error) { return []byte("encrypted:" + plaintext), nil }
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
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/admin/llm_config_management/... -v`
Expected: FAIL — `NewUsecaseImpl` undefined.

- [ ] **Step 4: Write the usecase implementation**

`backend/internal/application/admin/llm_config_management/usecase.go`:

```go
package applicationadminllmconfigmanagement

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
)

type usecase struct {
	repository domaincontractsrepository.LlmConfig
	encryptor  domaincontractsutility.Encryptor
	logger     domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	repository domaincontractsrepository.LlmConfig,
	encryptor domaincontractsutility.Encryptor,
	logger domaincontractslogger.Leveled,
) domainusecasesadmin.LlmConfigManagement {
	return &usecase{repository: repository, encryptor: encryptor, logger: logger}
}

func (u *usecase) Get(ctx context.Context) (*domainmodels.LlmConfig, error) {
	const tag = "admin/llm_config_management/Get"

	config, err := u.repository.Get(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read llm config", domainmodels.LoggerMeta{"err": err})
		return nil, err
	}
	return config, nil
}

func (u *usecase) Update(ctx context.Context, request domainusecasesadmin.UpdateLlmConfigRequest) error {
	const tag = "admin/llm_config_management/Update"

	provider, err := applicationshared.RequiredLlmProvider(request.Provider, "provider")
	if err != nil {
		return err
	}
	model, err := applicationshared.RequiredLlmModel(request.Model, "model")
	if err != nil {
		return err
	}
	apiKey, err := applicationshared.RequiredLlmApiKey(request.ApiKey, "api_key")
	if err != nil {
		return err
	}
	baseURL, err := applicationshared.OptionalLlmBaseURL(request.BaseURL, "base_url")
	if err != nil {
		return err
	}

	encryptedApiKey, err := u.encryptor.Encrypt(apiKey)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to encrypt api key", domainmodels.LoggerMeta{"err": err})
		return err
	}

	if _, err := u.repository.Upsert(ctx, provider, model, encryptedApiKey, baseURL); err != nil {
		u.logger.Error(ctx, tag, "failed to upsert llm config", domainmodels.LoggerMeta{"err": err})
		return err
	}
	return nil
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/admin/llm_config_management/... -v`
Expected: PASS on all three tests.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/usecases/admin/ internal/application/admin/llm_config_management/`
Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/usecases/admin/llm_config_management.go \
        backend/internal/application/admin/llm_config_management/
git commit -m "feat: add LlmConfigManagement usecase"
```

---

### Task 10: HTTP presentation layer

**Files:**
- Create: `backend/internal/presentation/http/response/llm_config.go`
- Modify: `backend/internal/presentation/http/request/admin.go`
- Modify: `backend/internal/presentation/http/handler/admin/handler.go`
- Modify: `backend/internal/presentation/http/route/route.go`
- Modify: `backend/database/seeder/permissions.json` (or the equivalent existing seed file — match whatever file already seeds `permission:get`/`permission:add`/etc.)

**Interfaces:**
- Consumes: `domainusecasesadmin.LlmConfigManagement` (Task 9), `domainmodels.LlmConfig` (Task 1).
- Produces: `presentationhttpresponse.LlmConfigResponse`, `presentationhttpresponse.LlmConfig(model domainmodels.LlmConfig) LlmConfigResponse`, `presentationhttprequest.LlmConfigPutRequest`, `handler.LlmConfigGet`/`LlmConfigPut` on the admin handler, `GET /v1/admin/llm-config` and `PUT /v1/admin/llm-config` routes gated by `llm_config:get` / `llm_config:set`.

There is no dedicated handler-level test in this codebase's convention (handlers are thin adapters over already-tested usecases) — verification is `go build`/`go vet`, matching every other handler method in `handler.go`.

- [ ] **Step 1: Write the redacted response type**

`backend/internal/presentation/http/response/llm_config.go`:

```go
package presentationhttpresponse

import (
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

// LlmConfigResponse never carries the API key in any form — not full, not
// partial. Only whether one is currently set.
type LlmConfigResponse struct {
	Provider     string     `json:"provider" example:"CLAUDE"`
	Model        string     `json:"model" example:"claude-opus-5"`
	BaseURL      *string    `json:"base_url,omitempty" example:"https://api.anthropic.com"`
	ApiKeySet    bool       `json:"api_key_set" example:"true"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

func LlmConfig(model *domainmodels.LlmConfig) LlmConfigResponse {
	if model == nil {
		return LlmConfigResponse{ApiKeySet: false}
	}
	updatedAt := model.UpdatedAt
	return LlmConfigResponse{
		Provider:  string(model.Provider),
		Model:     model.Model,
		BaseURL:   model.BaseURL,
		ApiKeySet: len(model.ApiKeyEncrypted) > 0,
		UpdatedAt: &updatedAt,
	}
}
```

- [ ] **Step 2: Add the request DTO**

Add to `backend/internal/presentation/http/request/admin.go`, in the same style as `PermissionPostRequest`:

```go
type LlmConfigPutRequest struct {
	Provider domainmodels.LlmProvider `json:"provider" example:"CLAUDE"`
	Model    string                   `json:"model" example:"claude-opus-5"`
	ApiKey   string                   `json:"api_key" example:"sk-ant-..."`
	BaseURL  *string                  `json:"base_url,omitempty" example:"https://api.anthropic.com"`
}
```

Add the `domainmodels` import to that file if it isn't already imported.

- [ ] **Step 3: Add the handler methods**

Add `llmConfigUseCase domainusecasesadmin.LlmConfigManagement` as a new field on the `handler` struct in `backend/internal/presentation/http/handler/admin/handler.go`, add it as the last positional parameter to `NewHandler(...)`, and add these two methods (following the exact swagger-comment and error-handling shape of `PermissionPatch`):

```go
// LlmConfigGet godoc
//
// @Summary LLM Config
// @Tags Admin - LLM Config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.LlmConfigResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/llm-config [get]
func (h *handler) LlmConfigGet(c *echo.Context) error {
	config, err := h.llmConfigUseCase.Get(c.Request().Context())
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	return c.JSON(http.StatusOK, presentationhttpresponse.LlmConfig(config))
}

// LlmConfigPut godoc
//
// @Summary LLM Config
// @Tags Admin - LLM Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.LlmConfigPutRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/llm-config [put]
func (h *handler) LlmConfigPut(c *echo.Context) error {
	var req presentationhttprequest.LlmConfigPutRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.llmConfigUseCase.Update(c.Request().Context(), domainusecasesadmin.UpdateLlmConfigRequest{
		Provider: req.Provider,
		Model:    req.Model,
		ApiKey:   req.ApiKey,
		BaseURL:  req.BaseURL,
	}); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
```

- [ ] **Step 4: Register the routes**

In `backend/internal/presentation/http/route/route.go`, add `LlmConfigGet(c *echo.Context) error` and `LlmConfigPut(c *echo.Context) error` to the `AdminHandler` interface, and add to `routeAdmin`:

```go
v1.GET("/admin/llm-config", handler.LlmConfigGet, permission("llm_config:get"))
v1.PUT("/admin/llm-config", handler.LlmConfigPut, permission("llm_config:set"))
```

- [ ] **Step 5: Seed the two new permissions**

In `backend/database/seeder/permissions.json` (or whichever file already seeds `permission:get` etc. — match its exact existing entry shape), add:

```json
{ "name": "llm_config:get", "description": "View the active LLM provider configuration." },
{ "name": "llm_config:set", "description": "Change the active LLM provider, model, or credential." }
```

and grant both to whichever role currently receives `permission:add`/`permission:set` (the admin role) in the corresponding role-permission seed file.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/presentation/`
Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/presentation/http/response/llm_config.go \
        backend/internal/presentation/http/request/admin.go \
        backend/internal/presentation/http/handler/admin/handler.go \
        backend/internal/presentation/http/route/route.go \
        backend/database/seeder/permissions.json
git commit -m "feat: add admin LLM config HTTP endpoints"
```

---

### Task 11: Composition wiring

**Files:**
- Modify: `backend/internal/composition/main/infrastructure.go`
- Modify: `backend/internal/composition/main/application.go`
- Modify: `backend/internal/composition/main/presentation.go`

**Interfaces:**
- Consumes: everything produced by Tasks 1–10.
- Produces: a fully running backend exposing `GET`/`PUT /v1/admin/llm-config`, with the `llm.ClientFactory` available on the `infrastructure` struct for later features (the IR recording plan) to consume.

- [ ] **Step 1: Wire the repository and encryption utility**

In `backend/internal/composition/main/infrastructure.go`, add fields to the `infrastructure` struct and construction lines to `newInfrastructure`, following the existing `permissionRepository`/`password` pattern:

```go
llmConfigRepository domaincontractsrepository.LlmConfig
llmEncryptor         domaincontractsutility.Encryptor
llmClientFactory      *infrastructurellm.ClientFactory
```

```go
llmConfigRepository := infrastructurerepositoryllmconfig.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)

llmEncryptor, err := infrastructureutilityencryption.NewAESGCMImpl(config.LlmEncryptionKey)
if err != nil {
	return nil, fmt.Errorf("failed to construct llm encryptor: %w", err)
}

llmClientFactory := infrastructurellm.NewClientFactory(llmConfigRepository, llmEncryptor)
```

Add the corresponding imports (`infrastructurerepositoryllmconfig`, `infrastructureutilityencryption`, `infrastructurellm`, `domaincontractsrepository` if not already imported, `domaincontractsutility` if not already imported).

- [ ] **Step 2: Wire the usecase**

In `backend/internal/composition/main/application.go`, add to the `application` struct and its constructor:

```go
adminLlmConfigManagement := applicationadminllmconfigmanagement.NewUsecaseImpl(
	l.infra.llmConfigRepository,
	l.infra.llmEncryptor,
	l.infra.logger,
)
```

- [ ] **Step 3: Wire the handler**

In `backend/internal/composition/main/presentation.go`, add `l.app.adminLlmConfigManagement` as the new final positional argument to the existing `presentationhttphandleradmin.NewHandler(...)` call.

- [ ] **Step 4: Verify end-to-end build**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l . && go test -count=1 ./...`
Expected: no build/vet/gofmt output; all tests pass (this runs every test written in Tasks 1–10 together for the first time).

- [ ] **Step 5: Smoke-test the endpoint**

With the backend running locally against a migrated Postgres and `BE_LLM_ENCRYPTION_KEY` set to a 32-byte value:

```bash
curl -X PUT http://127.0.0.1:8080/v1/admin/llm-config \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"provider":"CLAUDE","model":"claude-opus-5","api_key":"sk-ant-test-key"}'
# Expected: 204 No Content

curl http://127.0.0.1:8080/v1/admin/llm-config -H "Authorization: Bearer $ADMIN_TOKEN"
# Expected: 200 {"provider":"CLAUDE","model":"claude-opus-5","api_key_set":true,"updated_at":"..."}
# — no api_key field anywhere in the response.
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/composition/main/infrastructure.go \
        backend/internal/composition/main/application.go \
        backend/internal/composition/main/presentation.go
git commit -m "feat: wire LLM provider abstraction into composition root"
```

---

## Self-review notes

- **Spec coverage:** every element of the "LLM provider abstraction" section of `docs/superpowers/specs/2026-08-09-infrared-recording-session-design.md` has a task — contract (Task 5), Claude/OpenAI adapters (Tasks 6–7), per-call resolution not a startup singleton (Task 8), DB-backed config with admin read/write (Tasks 1–2, 9–10), never returning the key (Task 10 response type + explicit assertion in Step 5's smoke test).
- **Placeholder scan:** no `TBD`/`fill in`/bare "add error handling" steps. The two explicit "if the compiler rejects this type name" notes (Tasks 6 and 7) are not placeholders — they give real code plus a concrete, mechanical resolution procedure (grep the vendored SDK source, matching this skill's own guidance for statically-typed SDKs whose exact param-union names aren't pre-confirmed), and the request/response *shape* they depend on is independently pinned down by each task's own test.
- **Type consistency:** `domaincontractsllm.Client`/`GenerateTextRequest`/`GenerateTextResult`/`Usage` (Task 5) are used with identical field names in Tasks 6, 7, 8. `domainmodels.LlmConfig`/`LlmProvider` (Task 1) match across Tasks 2, 8, 9, 10. `ClientFactory.newClaudeClient`/`newOpenAIClient` field names in Task 8's test match the struct fields defined in Task 8's implementation step.
