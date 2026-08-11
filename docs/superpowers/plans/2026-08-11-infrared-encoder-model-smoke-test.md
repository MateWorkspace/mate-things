# Infrared Encoder-Generation Model Smoke Test Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Run the real `WriteCoder()` + real goja smoke-test gate against 9 real models (3 Claude, 3 OpenAI — both via OpenRouter — and 3 Gemini — via Google's native API) using a synthetic Polytron PAC-09HDN AC-remote fixture, and report pass/fail + latency + tokens per model.

**Architecture:** A single temporary, disposable Go program at
`backend/cmd/manual_infrared_coder_smoke/main.go` that is `go run` directly
(never committed, never added to the module's permanent command set). It
builds the fixture once, then loops over a hardcoded model list, calling
production code (`applicationinfraredcodergeneration.WriteCoder`,
`infrastructurellmclient.NewOpenAIClient`/`NewGeminiClient`,
`infrastructureutilityjsengine`'s goja `RunEncoder`) for each model, and
prints a results table at the end. The file is deleted as the final step of
this plan, per this session's established manual-smoke-test convention (see
`internal/infrastructure/llm/client/manual_gemini_real_smoke_test.go` from
earlier this session, already deleted).

**Tech Stack:** Go 1.x (existing backend module), existing
`domaincontractsllm.Client`/`GenerateTextRequest` contracts, existing
`infrastructurellmclient.NewOpenAIClient`/`NewGeminiClient` constructors,
existing `applicationinfraredcodergeneration.WriteCoder`, existing
`domaincontractsutility.JSEngine` / `infrastructureutilityjsengine.New()`
(goja-backed).

## Global Constraints

- Credentials come only from `openrouter.creds` and `googleai.creds` at the
  repo root (`/home/dodol/Repositories/mate/mate-things/`), read at program
  start — never hardcoded, never committed.
- The temporary program itself is never committed — it is `go run` directly
  from the working tree and deleted once the run's output has been captured.
- Each model call must not block the others: wrap each model's
  `WriteCoder` + `RunEncoder` pair in a 90-second timeout
  (`context.WithTimeout`), record a failure with that timeout's error on
  expiry, and move to the next model.
- Pass bar (from the design doc): `WriteCoder` returns without error (i.e.
  the model produced valid JSON matching the schema) AND
  `RunEncoder(encoderSource, baselineState, 2*time.Second)` returns a
  non-empty `[]int32` without erroring. Nothing else is graded.
- Generated `encoder_source`/`decoder_source` per model are written to
  `/tmp/claude-1000/-home-dodol-Repositories-mate/748a5526-9e98-4ad6-a85d-7bef7fe6c896/scratchpad/infrared-coder-smoke/<safe-model-name>.encoder.js`
  and `.decoder.js` (scratchpad dir per this session's system prompt) for
  manual inspection — written even on `RunEncoder` failure (only skipped if
  `WriteCoder` itself errored, since there's nothing to write).

---

### Task 1: Fixture + model list + report types

**Files:**
- Create: `backend/cmd/manual_infrared_coder_smoke/main.go`

**Interfaces:**
- Consumes:
  - `applicationinfraredanalysis.AnalysisPayload{FrameBitLength int, BaselineBits []int, ChecksumBits []int, VolatileBits []int, States []StateAttribution}`
  - `applicationinfraredanalysis.StateAttribution{StateId uuid.UUID, BitOffsets []int, ValueBits map[string][]int}`
  - `domainmodels.InfraredState{Id uuid.UUID, Name string, Type domainmodels.InfraredStateType}` (type `domainmodels.InfraredStateTypeEnum` for enum states, `domainmodels.InfraredStateTypeRange` for range states — confirmed present in `internal/domain/models/infrared.go`)
- Produces (for Task 2 to consume):
  - `type modelSpec struct { Provider string; Label string; Model string }`
  - `var models []modelSpec` — the 9 models below
  - `func buildFixture() (states []domainmodels.InfraredState, payload applicationinfraredanalysis.AnalysisPayload, baselineState map[string]string)`
  - `type result struct { Provider, Label, Model string; Latency time.Duration; InputTokens, OutputTokens int32; Passed bool; Err string }`

- [ ] **Step 1: Scaffold the file and package**

Create `backend/cmd/manual_infrared_coder_smoke/main.go`:

```go
// Command manual_infrared_coder_smoke is a disposable, one-off verification
// script — not part of the build, never committed. It exercises the real
// WriteCoder() encoder-generation path against real models from OpenRouter
// (Claude, OpenAI) and Google's Gemini API, using a synthetic AC-remote
// fixture, and reports pass/fail against the same goja smoke test
// production uses before persisting a coder.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	applicationinfraredanalysis "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/analysis"
	applicationinfraredcodergeneration "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/coder_generation"
	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurellmclient "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/llm/client"
	infrastructureutilityjsengine "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/utility/jsengine"
	"github.com/google/uuid"
)

const (
	repoRoot     = "/home/dodol/Repositories/mate/mate-things"
	scratchDir   = "/tmp/claude-1000/-home-dodol-Repositories-mate/748a5526-9e98-4ad6-a85d-7bef7fe6c896/scratchpad/infrared-coder-smoke"
	perModelTimeout = 90 * time.Second
	encoderSmokeTimeout = 2 * time.Second
)

type modelSpec struct {
	Provider string
	Label    string
	Model    string
}

var models = []modelSpec{
	{Provider: "openrouter-claude", Label: "Claude Opus 5", Model: "anthropic/claude-opus-5"},
	{Provider: "openrouter-claude", Label: "Claude Sonnet 5", Model: "anthropic/claude-sonnet-5"},
	{Provider: "openrouter-claude", Label: "Claude Haiku 4.5", Model: "anthropic/claude-haiku-4.5"},
	{Provider: "openrouter-openai", Label: "GPT-5", Model: "openai/gpt-5"},
	{Provider: "openrouter-openai", Label: "GPT-5 Mini", Model: "openai/gpt-5-mini"},
	{Provider: "openrouter-openai", Label: "GPT-4o", Model: "openai/gpt-4o"},
	{Provider: "gemini", Label: "Gemini Pro (latest)", Model: "gemini-pro-latest"},
	{Provider: "gemini", Label: "Gemini Flash (latest)", Model: "gemini-flash-latest"},
	{Provider: "gemini", Label: "Gemini Flash Lite (latest)", Model: "gemini-flash-lite-latest"},
}

type result struct {
	Provider     string
	Label        string
	Model        string
	Latency      time.Duration
	InputTokens  int32
	OutputTokens int32
	Passed       bool
	Err          string
}
```

- [ ] **Step 2: Add the fixture builder**

Append to the same file:

```go
// buildFixture returns a synthetic but realistic AnalysisPayload for a
// Polytron PAC-09HDN-style AC remote: 32-bit NEC-style frame, POWER at bit
// 0, MODE at bits 1-2, TEMP at bits 3-6, an 8-bit XOR checksum at bits
// 24-31, no volatile bits. Mirrors what buildAnalysisPayload produces from
// a real recording so WriteCoder's prompt reads naturally.
func buildFixture() ([]domainmodels.InfraredState, applicationinfraredanalysis.AnalysisPayload, map[string]string) {
	powerId := uuid.New()
	modeId := uuid.New()
	tempId := uuid.New()

	states := []domainmodels.InfraredState{
		{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
		{Id: tempId, Name: "TEMP", Type: domainmodels.InfraredStateTypeRange},
	}

	checksumBits := []int{24, 25, 26, 27, 28, 29, 30, 31}

	payload := applicationinfraredanalysis.AnalysisPayload{
		FrameBitLength: 32,
		BaselineBits:   make([]int, 32), // all zero: POWER=OFF, MODE=COOL, TEMP=24 baseline
		ChecksumBits:   checksumBits,
		VolatileBits:   nil,
		States: []applicationinfraredanalysis.StateAttribution{
			{
				StateId:    powerId,
				BitOffsets: []int{0},
				ValueBits: map[string][]int{
					"ON":  {0},
					"OFF": {},
				},
			},
			{
				StateId:    modeId,
				BitOffsets: []int{1, 2},
				ValueBits: map[string][]int{
					"COOL": {},
					"HEAT": {1},
					"FAN":  {2},
					"DRY":  {1, 2},
				},
			},
			{
				StateId:    tempId,
				BitOffsets: []int{3, 4, 5, 6},
				ValueBits: map[string][]int{
					"16": {3, 4, 5, 6},
					"24": {},
					"30": {4, 6},
				},
			},
		},
	}

	baselineState := map[string]string{
		"POWER": "OFF",
		"MODE":  "COOL",
		"TEMP":  "24",
	}

	return states, payload, baselineState
}
```

- [ ] **Step 3: Add a stub `main` so the file builds**

```go
func main() {
	states, payload, baselineState := buildFixture()
	fmt.Printf("fixture ready: %d states, %d-bit frame, baseline=%v\n", len(states), payload.FrameBitLength, baselineState)
}
```

- [ ] **Step 4: Verify it builds and runs**

Run: `cd backend && go run ./cmd/manual_infrared_coder_smoke`
Expected: prints `fixture ready: 3 states, 32-bit frame, baseline=map[MODE:COOL POWER:OFF TEMP:24]` (map key order may vary), exit 0, no compile errors.

- [ ] **Step 5: Commit is not applicable**

This file is disposable per the Global Constraints — do not `git add`/commit it at any point in this plan. Skip commit steps for every task; run `git status --short backend/cmd/` after Task 4 instead to confirm it's untracked, and delete it in Task 4's final step.

---

### Task 2: Credential loading + client construction

**Files:**
- Modify: `backend/cmd/manual_infrared_coder_smoke/main.go`

**Interfaces:**
- Consumes: `infrastructurellmclient.NewOpenAIClient(apiKey string, baseURL *string, model string) domaincontractsllm.Client`, `infrastructurellmclient.NewGeminiClient(apiKey string, baseURL *string, model string) domaincontractsllm.Client` (both confirmed present in `backend/internal/infrastructure/llm/client/{openai,gemini}.go`)
- Produces: `func loadCreds(path string) (baseURL, apiKey string)`, `func clientFor(spec modelSpec, openrouterBaseURL, openrouterKey, geminiBaseURL, geminiKey string) domaincontractsllm.Client`

- [ ] **Step 1: Add credential loading**

Append to `main.go`:

```go
// loadCreds reads a two-line creds file (base URL on line 1, API key on
// line 2) — same format as openrouter.creds/googleai.creds used elsewhere
// this session.
func loadCreds(path string) (baseURL, apiKey string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read creds file %s: %v\n", path, err)
		os.Exit(1)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) < 2 {
		fmt.Fprintf(os.Stderr, "expected 2 lines (base_url, api_key) in %s, got %d\n", path, len(lines))
		os.Exit(1)
	}
	return strings.TrimSpace(lines[0]), strings.TrimSpace(lines[1])
}
```

- [ ] **Step 2: Add client construction by provider**

```go
func clientFor(spec modelSpec, openrouterBaseURL, openrouterKey, geminiBaseURL, geminiKey string) domaincontractsllm.Client {
	switch spec.Provider {
	case "openrouter-claude", "openrouter-openai":
		// OpenRouter is OpenAI-compatible (chat completions), not
		// compatible with Anthropic's native Messages API — so Claude
		// models are exercised through the OpenAI client too, using
		// OpenRouter's own "anthropic/..." model slugs.
		return infrastructurellmclient.NewOpenAIClient(openrouterKey, &openrouterBaseURL, spec.Model)
	case "gemini":
		return infrastructurellmclient.NewGeminiClient(geminiKey, &geminiBaseURL, spec.Model)
	default:
		fmt.Fprintf(os.Stderr, "unknown provider %q for model %q\n", spec.Provider, spec.Model)
		os.Exit(1)
		return nil
	}
}
```

- [ ] **Step 3: Wire creds loading into `main` and verify**

Replace the stub `main` from Task 1 Step 3 with:

```go
func main() {
	states, payload, baselineState := buildFixture()

	openrouterBaseURL, openrouterKey := loadCreds(filepath.Join(repoRoot, "openrouter.creds"))
	geminiBaseURL, geminiKey := loadCreds(filepath.Join(repoRoot, "googleai.creds"))

	fmt.Printf("fixture ready: %d states, %d-bit frame, baseline=%v\n", len(states), payload.FrameBitLength, baselineState)
	fmt.Printf("openrouter base=%s key=%s...\n", openrouterBaseURL, openrouterKey[:12])
	fmt.Printf("gemini base=%s key=%s...\n", geminiBaseURL, geminiKey[:12])

	client := clientFor(models[0], openrouterBaseURL, openrouterKey, geminiBaseURL, geminiKey)
	fmt.Printf("built client for %s: %T\n", models[0].Model, client)
}
```

Run: `cd backend && go run ./cmd/manual_infrared_coder_smoke`
Expected: prints the fixture line, both creds lines (key prefix only — never
print the full key), and `built client for anthropic/claude-opus-5: *infrastructurellmclient.openAIClient` (or the actual unexported type name — check output, don't guess), exit 0.

---

### Task 3: Per-model run loop + goja smoke test + results table

**Files:**
- Modify: `backend/cmd/manual_infrared_coder_smoke/main.go`

**Interfaces:**
- Consumes:
  - `applicationinfraredcodergeneration.WriteCoder(ctx context.Context, client domaincontractsllm.Client, deviceBrand, deviceModel string, states []domainmodels.InfraredState, definitions []domainmodels.InfraredStateDeviceDefinition, payload applicationinfraredanalysis.AnalysisPayload) (applicationinfraredcodergeneration.Coder, error)` where `Coder{EncoderSource, DecoderSource, SummaryReadme, DetailReadme string}`
  - `domaincontractsutility.JSEngine.RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error)`
  - `infrastructureutilityjsengine.New() domaincontractsutility.JSEngine` (constructor name confirmed by reading `backend/internal/infrastructure/utility/jsengine/goja.go` before writing this step — if the actual exported constructor differs, e.g. no zero-arg `New()`, use whatever it actually exports)
- Produces: `func runOne(ctx context.Context, spec modelSpec, engine domaincontractsutility.JSEngine, states []domainmodels.InfraredState, payload applicationinfraredanalysis.AnalysisPayload, baselineState map[string]string, client domaincontractsllm.Client) result`, `func printReport(results []result)`

Confirmed by reading `backend/internal/infrastructure/utility/jsengine/goja.go`: the
exported constructor is `func NewGojaImpl() domaincontractsutility.JSEngine` (not
`New()` — corrected here so later steps use the real name).

- [ ] **Step 1: Add `runOne`**

```go
// runOne calls the real WriteCoder against spec's client, then — on
// success — runs the real goja smoke test against baselineState, exactly
// the check production performs before persisting a coder. Errors from
// either step are captured in result.Err rather than propagated, so one
// model's failure doesn't stop the others.
func runOne(
	ctx context.Context,
	spec modelSpec,
	engine domaincontractsutility.JSEngine,
	states []domainmodels.InfraredState,
	payload applicationinfraredanalysis.AnalysisPayload,
	baselineState map[string]string,
	client domaincontractsllm.Client,
) result {
	start := time.Now()
	res := result{Provider: spec.Provider, Label: spec.Label, Model: spec.Model}

	callCtx, cancel := context.WithTimeout(ctx, perModelTimeout)
	defer cancel()

	coder, err := applicationinfraredcodergeneration.WriteCoder(callCtx, client, "Polytron", "PAC-09HDN", states, nil, payload)
	res.Latency = time.Since(start)
	if err != nil {
		res.Err = fmt.Sprintf("WriteCoder: %v", err)
		return res
	}

	writeScratchFile(spec, "encoder.js", coder.EncoderSource)
	writeScratchFile(spec, "decoder.js", coder.DecoderSource)

	if _, err := engine.RunEncoder(coder.EncoderSource, baselineState, encoderSmokeTimeout); err != nil {
		res.Err = fmt.Sprintf("RunEncoder: %v", err)
		return res
	}

	res.Passed = true
	return res
}

func writeScratchFile(spec modelSpec, suffix, content string) {
	if err := os.MkdirAll(scratchDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir scratch dir: %v\n", err)
		return
	}
	safeName := strings.NewReplacer("/", "_", ":", "_", ".", "_").Replace(spec.Model)
	path := filepath.Join(scratchDir, fmt.Sprintf("%s.%s", safeName, suffix))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write scratch file %s: %v\n", path, err)
	}
}
```

Note: token usage (`InputTokens`/`OutputTokens`) is not filled in here
because `applicationinfraredcodergeneration.WriteCoder` returns only a
`Coder` struct (`EncoderSource, DecoderSource, SummaryReadme, DetailReadme`)
— it does not surface `domaincontractsllm.GenerateTextResult.Usage`. Leave
`res.InputTokens`/`res.OutputTokens` at their zero value and note this
explicitly as `n/a` in the report (Step 3), rather than fabricating numbers
or modifying production code's return type for this disposable script.

- [ ] **Step 2: Add `printReport` and wire the full loop into `main`**

```go
func printReport(results []result) {
	fmt.Println()
	fmt.Println("provider            model                              latency    pass   error")
	fmt.Println(strings.Repeat("-", 100))
	for _, r := range results {
		status := "FAIL"
		if r.Passed {
			status = "PASS"
		}
		fmt.Printf("%-19s %-34s %-10s %-6s %s\n", r.Provider, r.Model, r.Latency.Round(time.Millisecond), status, r.Err)
	}
}
```

Replace `main` (from Task 2 Step 3) with:

```go
func main() {
	states, payload, baselineState := buildFixture()

	openrouterBaseURL, openrouterKey := loadCreds(filepath.Join(repoRoot, "openrouter.creds"))
	geminiBaseURL, geminiKey := loadCreds(filepath.Join(repoRoot, "googleai.creds"))

	engine := infrastructureutilityjsengine.NewGojaImpl()

	ctx := context.Background()
	results := make([]result, 0, len(models))
	for _, spec := range models {
		client := clientFor(spec, openrouterBaseURL, openrouterKey, geminiBaseURL, geminiKey)
		fmt.Printf("running %s (%s)...\n", spec.Label, spec.Model)
		r := runOne(ctx, spec, engine, states, payload, baselineState, client)
		results = append(results, r)
	}

	printReport(results)
	fmt.Printf("\ngenerated sources written to: %s\n", scratchDir)
}
```

- [ ] **Step 3: Build-check (no network calls yet)**

Run: `cd backend && go build ./cmd/manual_infrared_coder_smoke`
Expected: exit 0, no output. This only proves it compiles — Task 4 does the real run.

---

### Task 4: Real run against all 9 models, capture report, clean up

**Files:**
- Modify (temporarily, then delete): `backend/cmd/manual_infrared_coder_smoke/main.go`

**Interfaces:**
- Consumes: everything from Tasks 1-3
- Produces: the final results table (reported to the user in chat, not persisted as a repo artifact) and the scratch-dir `.js` files (already covered by Global Constraints' scratch path)

- [ ] **Step 1: Run the full smoke test**

Run: `cd backend && go run ./cmd/manual_infrared_coder_smoke`
Expected: 9 `running ...` lines followed by a results table with 9 rows.
This is a live run against real paid APIs (OpenRouter + Google) — it will
take real wall-clock time (up to `9 * 90s` worst case) and cost real
tokens across 9 models generating ~8192-token JSON responses each.

- [ ] **Step 2: Investigate any FAIL rows**

For any row with `Passed=false`, read its `Err` column. If it's a
`WriteCoder` error, the scratch `.js` files won't exist for that model (skip
them). If it's a `RunEncoder` error, read that model's
`<safe-model-name>.encoder.js` from the scratch dir to see what the model
actually generated and why goja rejected it — this is genuinely useful
signal (e.g. the model used a JS feature goja's ES5.1 subset doesn't
support, or nested state values incorrectly), not a bug in the harness. Do
not silently retry or fix — report the finding, matching this plan's Task 4
Step 3.

- [ ] **Step 3: Report results to the user**

Summarize the results table, notable errors from Step 2, and where the
scratch files live, as a normal chat response — no artifact/document
required for this ephemeral exercise per the design doc's scope.

- [ ] **Step 4: Delete the disposable program**

```bash
rm backend/cmd/manual_infrared_coder_smoke/main.go
rmdir backend/cmd/manual_infrared_coder_smoke
cd backend && git status --short cmd/
```

Expected: `git status --short cmd/` prints nothing (the directory never
existed as far as git is concerned, since it was never `git add`ed) —
confirms full cleanup, consistent with every other manual smoke test this
session.
