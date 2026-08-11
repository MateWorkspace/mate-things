# Infrared encoder-generation model smoke test

## Context

`WriteCoder()` (`backend/internal/application/infrared/coder_generation/llm_coder.go`)
is the production function that turns a deterministic bit-analysis payload
into a JS encoder/decoder pair via an LLM call with a structured JSON
response schema. Today it's only exercised in unit tests against a fake
client. This is a one-off verification exercise (not a shipped feature) to
confirm that a spread of real models, across all three supported providers,
can actually perform this specific code-generation task and produce an
encoder that passes the same goja smoke test production uses before
persisting a coder.

Live IR hardware (to record real raw captures via MQTT) is not available in
this environment, so the full Start → Record → Analyze → Generate → Test
pipeline is out of scope. This exercises the LLM code-generation step in
isolation, using the real client/usecase code paths.

## Fixture

A synthetic but realistic AC-remote payload, continuing the
`Polytron PAC-09HDN` device already used in `llm_coder_test.go`:

- 32-bit frame (NEC-style: header + data byte + checksum byte, doubled for margin)
- `POWER` (ON/OFF) — bit 0
- `MODE` (COOL/HEAT/FAN/DRY) — bits 1–2
- `TEMP` (16–30°C) — bits 3–6
- Checksum — bits 24–31 (XOR of bytes 0–2), reported via `ChecksumBits`
- No volatile bits

This mirrors what `buildAnalysisPayload` produces from a real recording, so
`WriteCoder`'s prompt reads naturally. `AnalysisPayload.States[].ValueBits`
carries a representative sample of observed values per state (not an
exhaustive enumeration), matching how OFAT case generation actually records
a limited set of cases per state in production.

Baseline state for the smoke test: `POWER=OFF, MODE=COOL, TEMP=24`.

## Models under test (~9)

- **Claude** (Opus/Sonnet/Haiku tier) — via OpenRouter
- **OpenAI** (gpt-5/gpt-5-mini/gpt-4o tier) — via OpenRouter
- **Gemini** (`gemini-pro-latest`, `gemini-flash-latest`, `gemini-flash-lite-latest`) — via Google's native API

OpenRouter is OpenAI-compatible (chat completions), not compatible with
Anthropic's native Messages API — so both Claude and OpenAI models are
exercised through `NewOpenAIClient` pointed at OpenRouter's base URL, using
OpenRouter's own model slugs (e.g. `anthropic/claude-...`, `openai/gpt-...`).
Exact slugs are resolved at run time against OpenRouter's `/models` listing
rather than guessed, since OpenRouter's naming doesn't match this app's
internal model-id aliases. Gemini is exercised through `NewGeminiClient`
directly against Google's API, using credentials from `googleai.creds`.

Credentials: `openrouter.creds` (base URL + key, line 1/line 2) and
`googleai.creds` (same shape), both at the repo root. Read only for this
run; never hardcoded into a permanent file.

## Mechanics

A temporary Go program at `backend/cmd/manual_infrared_coder_smoke/main.go`
(deleted after the run — same disposable-script convention used earlier this
session for the OpenRouter and Gemini connectivity smoke tests). For each
model:

1. Build the client (`NewOpenAIClient` or `NewGeminiClient`) with that
   model's id and the appropriate creds.
2. Call the real `WriteCoder(ctx, client, "Polytron", "PAC-09HDN", states,
   nil, payload)`.
3. On success, run the real `goja` `RunEncoder(encoderSource, baselineState,
   2*time.Second)` — the identical check `runAnalysisAndGeneration` performs
   before persisting a coder.
4. Record: provider, model, wall-clock latency, input/output token usage,
   pass/fail, and error text (from either step).
5. Write that model's `encoder_source` and `decoder_source` to a scratch
   file (`/tmp/claude-.../scratchpad/infrared-coder-smoke/<model>.js`) for
   manual inspection.

Each model call runs independently — one model's failure doesn't block
others. A per-call timeout (e.g. 90s) prevents one hung request from
stalling the whole run.

## Report

A results table printed at the end: model · provider · latency · tokens ·
pass/fail · error (if any). Plus the scratch-file paths for anyone who wants
to read the generated JS.

## Out of scope

- Live IR hardware / MQTT capture / the full recording pipeline
- DB, broadcaster, or repository wiring
- Decoder correctness or bit-level grading of the generated encoder's output
  (pass bar is: valid structured JSON response + encoder parses and runs
  without erroring against the baseline state, same as production's gate)
