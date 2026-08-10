# IR Recording Session B2 (Analysis + Encoder/Decoder Generation) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Take a fully-recorded IR session (every case `ACCEPTED`, built by Plan B1) through deterministic bit-level analysis and LLM-generated JavaScript encoder/decoder code, executed via `goja`, ending with a persisted `InfraredStateCoder` row in `UNVERIFIED` status. Hardware-verified activation (`TEST_CASES_GENERATING` → `TESTING` → `COMPLETED`) is explicitly out of scope — that is Plan B3.

**Architecture:** A pure, dependency-free Go package (`applicationinfraredanalysis`) turns raw IR captures into a structured bit-attribution payload (frame boundaries, volatile bits, checksum bits, per-state bit placement) using nothing but the OFAT structure Plan B1 already guarantees. That payload, plus device/state context, goes through one LLM call (`applicationinfraredcodergeneration.WriteCoder`, this plan's second consumer of the merged LLM provider abstraction, after Plan B1's `WriteScript`) requesting a structured JSON response containing the encoder/decoder source and two README fields. The generated encoder is smoke-tested once via a new `goja`-backed execution engine before being persisted — not for signal correctness (hardware is the verification authority, per the design spec's own framing), only to catch code that doesn't run at all. The whole chain (`ANALYZING` → `FUNCTION_GENERATING`) runs as a second background-goroutine job on the same `RecordSessionManagement` usecase Plan B1 built, triggered automatically the moment the last case's second raw capture is accepted — no new staff-facing "start analysis" action exists; staff only ever accepts/discards raws and retries cases, exactly as in B1.

**Tech Stack:** Go, `goja` (new dependency — a pure-Go ECMAScript 5.1+ interpreter, no cgo, no external process), the already-merged `domaincontractsllm.Client`/`infrastructurellm.ClientFactory` abstraction, PostgreSQL via the existing `BasePostgres`/squirrel pattern.

## Global Constraints

- `InfraredRecordSession.RecordingState` stays the plain `string` field from Plan B1 (not a typed enum) — this plan adds no new lifecycle stages beyond the ones Plan B1's Task 1 already defined (`ANALYZING`, `FUNCTION_GENERATING` already exist as untyped string constants).
- No new websocket broadcaster, no new MQTT topics. This plan reuses Plan B1's `domaincontractsbroadcaster.InfraredRecordSession`/`domainmodels.InfraredRecordSessionEvent` for all status updates and needs no wire-level additions to either.
- No `Status`-typed session field exists — every place below that needs to compare against a lifecycle stage compares `RecordingState` against the existing untyped `domainmodels.InfraredRecordingState*` constants from Plan B1, never a `Status`/enum type. (The original design spec used a typed `Status InfraredRecordSessionStatus` field; that was superseded by Plan B1's actual implementation and is not revived here — flagged explicitly since the spec document predates that decision.)
- `InfraredStateCoder.Status` (new in this plan) IS a typed string enum (`InfraredStateCoderStatus`), consistent with `InfraredRecordCaseStatus`/`InfraredRecordRawStatus`'s existing pattern — the "session lifecycle stage stays a plain string" decision from B1 applies only to the session's own pipeline-stage field, not to every status field in the domain.
- This plan's terminal state for a session is sitting in `RecordingState = FUNCTION_GENERATING` with exactly one `InfraredStateCoder` row in `UNVERIFIED` status persisted against it. The `FUNCTION_GENERATING → TEST_CASES_GENERATING` transition is Plan B3's first order of business, not this plan's — nothing here auto-advances past `FUNCTION_GENERATING`.
- No round-trip signal-correctness verification anywhere in this plan. The one `goja` execution this plan performs (smoke-testing the generated encoder against the baseline case's state) exists only to catch code that throws or returns a malformed result — never to validate that the produced IR timing is correct. Hardware is the verification authority (Plan B3).
- No new HTTP write endpoints beyond one read endpoint (`GET .../coder`). Nothing in this plan is staff-triggered; the whole analysis→generation chain is fully automatic, fired from the same `AcceptRaw` call staff already use in Plan B1.

---

### Task 1: Domain model — `InfraredStateCoder`

**Files:**
- Modify: `backend/internal/domain/models/infrared.go`

**Interfaces:**
- Produces: `domainmodels.InfraredStateCoderStatus` (`UNVERIFIED`/`ACTIVE`/`SUPERSEDED`) and `domainmodels.InfraredStateCoder{Id, InfraredDeviceId, InfraredRecordSessionId, EncoderSource, DecoderSource, SummaryReadme, DetailReadme, Status, CreatedAt}`.

This plan only ever creates a coder in `UNVERIFIED` status — `ACTIVE`/`SUPERSEDED` exist on the type now because they're part of the same enum's closed membership, but no code in this plan transitions to them (that's Plan B3's "all tests pass" step).

- [ ] **Step 1: Add the type**

Append to `backend/internal/domain/models/infrared.go`:

```go
type InfraredStateCoderStatus string

const (
	InfraredStateCoderStatusUnverified InfraredStateCoderStatus = "UNVERIFIED"
	InfraredStateCoderStatusActive     InfraredStateCoderStatus = "ACTIVE"
	InfraredStateCoderStatusSuperseded InfraredStateCoderStatus = "SUPERSEDED"
)

type InfraredStateCoder struct {
	Id                      uuid.UUID
	InfraredDeviceId        uuid.UUID
	InfraredRecordSessionId uuid.UUID
	EncoderSource           string
	DecoderSource           string
	SummaryReadme           string
	DetailReadme            string
	Status                  InfraredStateCoderStatus
	CreatedAt               time.Time
}
```

`time` and `uuid` are already imported in this file (Plan B1's `InfraredRecordSession`/`CreatedAt` uses both) — no new imports needed.

- [ ] **Step 2: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/models/`
Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/domain/models/infrared.go
git commit -m "feat: add InfraredStateCoder domain model"
```

---

### Task 2: Migration — `infrared_state_coder` table

**Files:**
- Create: `backend/database/migrations/20260810090002_infrared_state_coder.up.sql`
- Create: `backend/database/migrations/20260810090002_infrared_state_coder.down.sql`

**Interfaces:**
- Produces: `infrared_state_coder` table.

- [ ] **Step 1: Write the migration**

`backend/database/migrations/20260810090002_infrared_state_coder.up.sql`:

```sql
CREATE TABLE infrared_state_coder (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_device_id UUID NOT NULL REFERENCES infrared_device (id) ON DELETE CASCADE,
    infrared_record_session_id UUID NOT NULL REFERENCES infrared_record_session (id) ON DELETE CASCADE,
    encoder_source TEXT NOT NULL,
    decoder_source TEXT NOT NULL,
    summary_readme TEXT NOT NULL,
    detail_readme TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'UNVERIFIED',
    CONSTRAINT chk_infrared_state_coder_status CHECK (status IN ('UNVERIFIED', 'ACTIVE', 'SUPERSEDED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_infrared_state_coder_infrared_device_id ON infrared_state_coder (infrared_device_id);
CREATE INDEX idx_infrared_state_coder_infrared_record_session_id ON infrared_state_coder (infrared_record_session_id);
```

`backend/database/migrations/20260810090002_infrared_state_coder.down.sql`:

```sql
DROP TABLE IF EXISTS infrared_state_coder;
```

- [ ] **Step 2: Apply and verify**

Run: `migrate -path backend/database/migrations -database "$DATABASE_URL" up`
Expected: no error; `\d infrared_state_coder` in `psql` shows the columns above. If no database is reachable in this environment, skip this step and say so explicitly — do not fabricate output — but confirm the SQL's filename/format matches sibling migrations (`20260810090001_infrared_recording.*`) and that `infrared_device`/`infrared_record_session` (the two tables referenced) genuinely exist in an earlier migration.

- [ ] **Step 3: Verify build**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l database/`
Expected: no output (this task has no Go code).

- [ ] **Step 4: Commit**

```bash
git add backend/database/migrations/20260810090002_infrared_state_coder.*
git commit -m "feat: add infrared_state_coder table"
```

---

### Task 3: `InfraredStateCoder` repository

**Files:**
- Create: `backend/internal/domain/contracts/repository/infrared_state_coder.go`
- Create: `backend/internal/infrastructure/repository/infrared_state_coder/postgres.go`, `postgres_query.go`

**Interfaces:**
- Consumes: `domainmodels.InfraredStateCoder` (Task 1).
- Produces: `domaincontractsrepository.InfraredStateCoder` with `Create(ctx, coder domainmodels.InfraredStateCoder) (id uuid.UUID, err error)` and `GetBySessionId(ctx, sessionId uuid.UUID) (*domainmodels.InfraredStateCoder, error)`.

No `List`/`Update`/`Delete` in this plan — a session produces at most one coder in B2's scope (retry-driven multiple-coders-per-session is Plan B3's concern, once `RECORDING` can be re-entered from `TESTING`). No dedicated test, per this codebase's established repository-layer convention (matches every repository task in Plan B1 — none of `infrared_record_session`, `infrared_state_device_record_case`, etc. have a test file).

- [ ] **Step 1: Write the domain contract**

```go
// backend/internal/domain/contracts/repository/infrared_state_coder.go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredStateCoder interface {
	Create(ctx context.Context, coder domainmodels.InfraredStateCoder) (id uuid.UUID, err error)
	GetBySessionId(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredStateCoder, error)
}
```

`Create` takes the whole struct (rather than one parameter per field, like `InfraredRecordSession.Create`'s narrower signature) because this row has 6 meaningful fields at creation time — passing a struct avoids an unreadably long parameter list. `Status`/`Id`/`CreatedAt` in the passed struct are ignored by the implementation (the row always starts `UNVERIFIED` with a generated `id`/`created_at`); only `InfraredDeviceId`, `InfraredRecordSessionId`, `EncoderSource`, `DecoderSource`, `SummaryReadme`, `DetailReadme` are read from it.

- [ ] **Step 2: Write the Postgres implementation**

Same `BasePostgres`/squirrel pattern as every Plan B1 repository. Follow `infrared_record_session/postgres.go`'s `GetById`-style not-found handling (`errors.Is(err, pgx.ErrNoRows)` → `infrastructurerepositoryshared.NotFound(...)`) for `GetBySessionId`.

```go
// backend/internal/infrastructure/repository/infrared_state_coder/postgres.go
package infrastructurerepositoryinfraredstatecoder

import (
	"context"
	"errors"

	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(
	dt pgxdt.Pgxdt,
	sqrQuestion *squirrel.StatementBuilderType,
	sqrDollar *squirrel.StatementBuilderType,
) domaincontractsrepository.InfraredStateCoder {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) Create(ctx context.Context, coder domainmodels.InfraredStateCoder) (id uuid.UUID, err error) {
	query, args, err := p.SqrD.Insert("infrared_state_coder").
		Columns("infrared_device_id", "infrared_record_session_id", "encoder_source", "decoder_source", "summary_readme", "detail_readme").
		Values(coder.InfraredDeviceId, coder.InfraredRecordSessionId, coder.EncoderSource, coder.DecoderSource, coder.SummaryReadme, coder.DetailReadme).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_state_coder query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_state_coder", err)
	}
	return id, nil
}

func (p *postgresImpl) GetBySessionId(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	query, args, err := p.SqrD.
		Select("id", "infrared_device_id", "infrared_record_session_id", "encoder_source", "decoder_source", "summary_readme", "detail_readme", "status", "created_at").
		From("infrared_state_coder").
		Where(squirrel.Eq{"infrared_record_session_id": sessionId}).
		ToSql()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build get infrared_state_coder query", err)
	}

	var item domainmodels.InfraredStateCoder
	var status string
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(
		&item.Id, &item.InfraredDeviceId, &item.InfraredRecordSessionId,
		&item.EncoderSource, &item.DecoderSource, &item.SummaryReadme, &item.DetailReadme,
		&status, &item.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_state_coder not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to get infrared_state_coder", err)
	}
	item.Status = domainmodels.InfraredStateCoderStatus(status)
	return &item, nil
}
```

Move the two inline `Select(...).ToSql()`/`Insert(...).ToSql()` calls into `postgres_query.go` methods instead if the reviewer of this task's PR prefers matching the two-file split exactly (same latitude every Plan B1 repository task was given).

- [ ] **Step 3: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/infrared_state_coder/`
Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/contracts/repository/infrared_state_coder.go \
        backend/internal/infrastructure/repository/infrared_state_coder/
git commit -m "feat: add InfraredStateCoder repository"
```

---

### Task 4: Frame segmentation and bit demodulation

**Files:**
- Create: `backend/internal/application/infrared/analysis/frame.go`
- Create: `backend/internal/application/infrared/analysis/frame_test.go`

**Interfaces:**
- Produces: `applicationinfraredanalysis.SegmentFrames(raw []int32) [][]int32` and `applicationinfraredanalysis.DemodulateBits(frame []int32) []int`.

Pure logic, no I/O — this and the next two tasks are this plan's strongest TDD candidates, exactly like Plan B1's OFAT generator (Task 8 there).

**Design and stated assumptions (read before implementing):** `raw` is a flat sequence of alternating mark/space durations in microseconds (mark at even indices, space at odd indices — matches the MQTT `ir_capture` payload shape Plan B1 already defined). Frame segmentation splits on any space longer than `frameGapThresholdMicros` (5000µs), which distinguishes an inter-frame gap from a normal data-bit space for the great majority of consumer IR protocols (NEC-family data-bit spaces top out around 1690–2250µs). Bit demodulation treats the FIRST mark+space pair of a frame as a header/preamble and excludes it from the returned bits, then classifies every remaining space as `0` or `1` by comparing it against the midpoint between the frame's minimum and maximum space duration (an adaptive threshold, not a hardcoded microsecond value, so it isn't tied to one manufacturer's exact timing). **This is a deliberate, documented simplification** — some protocols have no distinguishable header, and the midpoint heuristic can misclassify on a frame with very few bits. Both are accepted limitations for this plan's first version, not blockers.

- [ ] **Step 1: Write the failing tests**

`backend/internal/application/infrared/analysis/frame_test.go`:

```go
package applicationinfraredanalysis

import (
	"reflect"
	"testing"
)

func TestSegmentFramesSplitsOnLongGap(t *testing.T) {
	// header(9000,4500) + 2 data bits (560,560=0) (560,1690=1) + long gap(20000) + repeat frame
	raw := []int32{9000, 4500, 560, 560, 560, 1690, 20000, 9000, 4500, 560, 560}
	frames := SegmentFrames(raw)
	if len(frames) != 2 {
		t.Fatalf("SegmentFrames() returned %d frames, want 2", len(frames))
	}
	if !reflect.DeepEqual(frames[0], []int32{9000, 4500, 560, 560, 560, 1690}) {
		t.Errorf("frames[0] = %v, want the first 6 values", frames[0])
	}
	if !reflect.DeepEqual(frames[1], []int32{9000, 4500, 560, 560}) {
		t.Errorf("frames[1] = %v, want the repeat frame", frames[1])
	}
}

func TestSegmentFramesSingleFrameWhenNoLongGap(t *testing.T) {
	raw := []int32{9000, 4500, 560, 560, 560, 1690}
	frames := SegmentFrames(raw)
	if len(frames) != 1 {
		t.Fatalf("SegmentFrames() returned %d frames, want 1", len(frames))
	}
}

func TestDemodulateBitsSkipsHeaderAndClassifiesBySpaceMidpoint(t *testing.T) {
	// header(9000,4500) then bits: short space=0, long space=1, short=0, long=1
	frame := []int32{9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 1690}
	bits := DemodulateBits(frame)
	want := []int{0, 1, 0, 1}
	if !reflect.DeepEqual(bits, want) {
		t.Fatalf("DemodulateBits() = %v, want %v", bits, want)
	}
}

func TestDemodulateBitsIgnoresDanglingTrailingMark(t *testing.T) {
	// header + one bit + a trailing mark with no paired space
	frame := []int32{9000, 4500, 560, 560, 560}
	bits := DemodulateBits(frame)
	if !reflect.DeepEqual(bits, []int{0}) {
		t.Fatalf("DemodulateBits() = %v, want [0] (trailing unpaired mark ignored)", bits)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/analysis/... -run "TestSegmentFrames|TestDemodulateBits" -v`
Expected: FAIL — `SegmentFrames`/`DemodulateBits` undefined.

- [ ] **Step 3: Write the implementation**

`backend/internal/application/infrared/analysis/frame.go`:

```go
package applicationinfraredanalysis

const frameGapThresholdMicros = 5000

// SegmentFrames splits a flat mark/space duration sequence into candidate
// frames, treating any space longer than frameGapThresholdMicros as an
// inter-frame gap rather than a data bit. See this task's design note in
// the implementation plan for why this threshold and not a per-protocol one.
func SegmentFrames(raw []int32) [][]int32 {
	if len(raw) == 0 {
		return nil
	}

	var frames [][]int32
	start := 0
	for i := 1; i < len(raw); i += 2 {
		if raw[i] > frameGapThresholdMicros {
			frames = append(frames, raw[start:i])
			start = i + 1
		}
	}
	frames = append(frames, raw[start:])
	return frames
}

// DemodulateBits converts one frame's mark/space durations into a bit
// sequence, treating the first mark+space pair as a header (excluded from
// the result) and classifying every subsequent space against the midpoint
// of that frame's own minimum and maximum space duration.
func DemodulateBits(frame []int32) []int {
	if len(frame) < 4 {
		return nil
	}

	spaces := make([]int32, 0, len(frame)/2)
	for i := 3; i < len(frame); i += 2 {
		spaces = append(spaces, frame[i])
	}
	if len(spaces) == 0 {
		return nil
	}

	min, max := spaces[0], spaces[0]
	for _, s := range spaces {
		if s < min {
			min = s
		}
		if s > max {
			max = s
		}
	}
	midpoint := (min + max) / 2

	bits := make([]int, len(spaces))
	for i, s := range spaces {
		if s > midpoint {
			bits[i] = 1
		}
	}
	return bits
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/analysis/... -v`
Expected: PASS on all four tests.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/infrared/analysis/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/infrared/analysis/frame.go backend/internal/application/infrared/analysis/frame_test.go
git commit -m "feat: add IR frame segmentation and bit demodulation"
```

---

### Task 5: Volatile-bit detection

**Files:**
- Create: `backend/internal/application/infrared/analysis/volatile.go`
- Create: `backend/internal/application/infrared/analysis/volatile_test.go`

**Interfaces:**
- Produces: `applicationinfraredanalysis.DetectVolatileBits(a, b []int) (map[int]struct{}, error)` and `applicationinfraredanalysis.UnionVolatileBits(sets ...map[int]struct{}) map[int]struct{}`.

Pure logic, no I/O. `DetectVolatileBits` compares two same-length bit sequences from the *same case's* two accepted captures — any index where they differ is volatile (a rolling checksum/timestamp/toggle bit, not a semantically meaningful state bit). A length mismatch between the two captures of one case is treated as a hard error (something is wrong with the capture, not a normal volatile bit) rather than silently ignored — this propagates up to a `FAILED` transition in Task 10's usecase wiring, consistent with every other analysis-phase failure path.

- [ ] **Step 1: Write the failing tests**

`backend/internal/application/infrared/analysis/volatile_test.go`:

```go
package applicationinfraredanalysis

import (
	"errors"
	"reflect"
	"testing"
)

func TestDetectVolatileBitsFindsDifferingPositions(t *testing.T) {
	a := []int{0, 1, 0, 1, 0}
	b := []int{0, 1, 1, 1, 1}
	got, err := DetectVolatileBits(a, b)
	if err != nil {
		t.Fatalf("DetectVolatileBits() error = %v, want nil", err)
	}
	want := map[int]struct{}{2: {}, 4: {}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DetectVolatileBits() = %v, want %v", got, want)
	}
}

func TestDetectVolatileBitsReturnsEmptySetWhenIdentical(t *testing.T) {
	a := []int{1, 0, 1}
	got, err := DetectVolatileBits(a, append([]int{}, a...))
	if err != nil {
		t.Fatalf("DetectVolatileBits() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Fatalf("DetectVolatileBits() = %v, want empty", got)
	}
}

func TestDetectVolatileBitsErrorsOnLengthMismatch(t *testing.T) {
	_, err := DetectVolatileBits([]int{0, 1}, []int{0, 1, 1})
	if err == nil {
		t.Fatal("DetectVolatileBits() error = nil, want a length-mismatch error")
	}
}

func TestUnionVolatileBitsCombinesMultipleSets(t *testing.T) {
	got := UnionVolatileBits(
		map[int]struct{}{1: {}},
		map[int]struct{}{2: {}, 3: {}},
		map[int]struct{}{1: {}},
	)
	want := map[int]struct{}{1: {}, 2: {}, 3: {}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("UnionVolatileBits() = %v, want %v", got, want)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/analysis/... -run "TestDetectVolatileBits|TestUnionVolatileBits" -v`
Expected: FAIL — undefined functions.

- [ ] **Step 3: Write the implementation**

`backend/internal/application/infrared/analysis/volatile.go`:

```go
package applicationinfraredanalysis

import "fmt"

// DetectVolatileBits compares two bit sequences captured for the SAME case
// and returns the set of positions where they differ — these are volatile
// (rolling checksum, timestamp, toggle bit), not semantically meaningful.
func DetectVolatileBits(a, b []int) (map[int]struct{}, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("bit sequence length mismatch: %d vs %d", len(a), len(b))
	}

	volatile := make(map[int]struct{})
	for i := range a {
		if a[i] != b[i] {
			volatile[i] = struct{}{}
		}
	}
	return volatile, nil
}

// UnionVolatileBits combines volatile-bit sets detected across multiple
// cases into one mask, since a bit volatile in any case must be excluded
// from semantic analysis everywhere.
func UnionVolatileBits(sets ...map[int]struct{}) map[int]struct{} {
	union := make(map[int]struct{})
	for _, set := range sets {
		for bit := range set {
			union[bit] = struct{}{}
		}
	}
	return union
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/analysis/... -v`
Expected: PASS on every test in the package so far.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/infrared/analysis/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/infrared/analysis/volatile.go backend/internal/application/infrared/analysis/volatile_test.go
git commit -m "feat: add IR volatile-bit detection"
```

---

### Task 6: Checksum discovery and per-state bit attribution

**Files:**
- Create: `backend/internal/application/infrared/analysis/attribution.go`
- Create: `backend/internal/application/infrared/analysis/attribution_test.go`

**Interfaces:**
- Consumes: `applicationinfraredanalysis`'s own types from Tasks 4-5 (bit sequences, volatile sets) — takes `uuid.UUID` as an opaque key type but no domain-model or repository types, keeping this package fully decoupled from persistence (the usecase in Task 10 is responsible for turning repository rows into the plain maps this function takes).
- Produces: `applicationinfraredanalysis.StateAttribution{StateId uuid.UUID, BitOffsets []int, ValueBits map[string][]int}` and `applicationinfraredanalysis.AnalysisPayload{FrameBitLength int, BaselineBits []int, ChecksumBits []int, VolatileBits []int, States []StateAttribution}`, produced by `applicationinfraredanalysis.Attribute(baselineBits []int, caseBits map[uuid.UUID][]int, caseTargetState map[uuid.UUID]uuid.UUID, caseTargetValue map[uuid.UUID]string, volatile map[int]struct{}) (AnalysisPayload, error)`.

This is the core algorithm the design spec calls "checksum discovery via intersection of per-state deltas" and "per-state bit attribution." Given the one-factor-at-a-time structure Plan B1's case generator guarantees (every non-baseline case differs from baseline in exactly one state), a bit position is a genuine checksum candidate if and only if it shows up in the delta-set of cases belonging to **more than one different state** — a bit truly owned by one state should only ever change when that state's value changes, never as a side effect of some other state changing. `caseTargetState`/`caseTargetValue` are precomputed by the usecase from the already-persisted `InfraredStateDeviceRecordState` rows (each non-baseline case differs from baseline in exactly one state by construction, so "which state does this case target" is a direct lookup, not something this pure package infers).

- [ ] **Step 1: Write the failing tests**

`backend/internal/application/infrared/analysis/attribution_test.go`:

```go
package applicationinfraredanalysis

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
)

func TestAttributeAssignsNonOverlappingBitsToEachState(t *testing.T) {
	powerId, modeId := uuid.New(), uuid.New()

	baseline := []int{0, 0, 0, 0} // POWER at bit0, MODE at bits1-2, bit3 unused
	caseBits := map[uuid.UUID][]int{
		powerId: {1, 0, 0, 0}, // POWER=OFF flips only bit0
		modeId:  {0, 1, 0, 0}, // MODE=HEAT flips only bit1
	}
	caseTargetState := map[uuid.UUID]uuid.UUID{powerId: powerId, modeId: modeId}
	caseTargetValue := map[uuid.UUID]string{powerId: "OFF", modeId: "HEAT"}

	payload, err := Attribute(baseline, caseBits, caseTargetState, caseTargetValue, map[int]struct{}{})
	if err != nil {
		t.Fatalf("Attribute() error = %v, want nil", err)
	}
	if len(payload.ChecksumBits) != 0 {
		t.Fatalf("ChecksumBits = %v, want empty (no bit changed for more than one state)", payload.ChecksumBits)
	}

	byState := make(map[uuid.UUID]StateAttribution, len(payload.States))
	for _, s := range payload.States {
		byState[s.StateId] = s
	}
	if !reflect.DeepEqual(byState[powerId].BitOffsets, []int{0}) {
		t.Errorf("POWER bit offsets = %v, want [0]", byState[powerId].BitOffsets)
	}
	if !reflect.DeepEqual(byState[modeId].BitOffsets, []int{1}) {
		t.Errorf("MODE bit offsets = %v, want [1]", byState[modeId].BitOffsets)
	}
	if !reflect.DeepEqual(byState[powerId].ValueBits["OFF"], []int{1}) {
		t.Errorf("POWER OFF value bits = %v, want [1]", byState[powerId].ValueBits["OFF"])
	}
}

func TestAttributeIdentifiesChecksumBitSharedAcrossStates(t *testing.T) {
	powerId, modeId := uuid.New(), uuid.New()

	baseline := []int{0, 0, 0, 0} // bit3 = checksum, flips whenever anything changes
	caseBits := map[uuid.UUID][]int{
		powerId: {1, 0, 0, 1}, // POWER=OFF: bit0 (owned) + bit3 (checksum)
		modeId:  {0, 1, 0, 1}, // MODE=HEAT: bit1 (owned) + bit3 (checksum)
	}
	caseTargetState := map[uuid.UUID]uuid.UUID{powerId: powerId, modeId: modeId}
	caseTargetValue := map[uuid.UUID]string{powerId: "OFF", modeId: "HEAT"}

	payload, err := Attribute(baseline, caseBits, caseTargetState, caseTargetValue, map[int]struct{}{})
	if err != nil {
		t.Fatalf("Attribute() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(payload.ChecksumBits, []int{3}) {
		t.Fatalf("ChecksumBits = %v, want [3]", payload.ChecksumBits)
	}

	for _, s := range payload.States {
		for _, offset := range s.BitOffsets {
			if offset == 3 {
				t.Fatalf("state %v owns bit 3, which is a checksum bit and must be excluded", s.StateId)
			}
		}
	}
}

func TestAttributeExcludesVolatileBitsFromEverything(t *testing.T) {
	powerId := uuid.New()
	baseline := []int{0, 0, 0}
	caseBits := map[uuid.UUID][]int{powerId: {1, 0, 1}} // bit2 volatile, would otherwise look checksum-like
	caseTargetState := map[uuid.UUID]uuid.UUID{powerId: powerId}
	caseTargetValue := map[uuid.UUID]string{powerId: "OFF"}
	volatile := map[int]struct{}{2: {}}

	payload, err := Attribute(baseline, caseBits, caseTargetState, caseTargetValue, volatile)
	if err != nil {
		t.Fatalf("Attribute() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(payload.VolatileBits, []int{2}) {
		t.Fatalf("VolatileBits = %v, want [2]", payload.VolatileBits)
	}
	for _, s := range payload.States {
		for _, offset := range s.BitOffsets {
			if offset == 2 {
				t.Fatalf("state %v owns volatile bit 2, must be excluded", s.StateId)
			}
		}
	}
	sort.Ints(payload.ChecksumBits) // no-op if empty, keeps this test order-independent
	if len(payload.ChecksumBits) != 0 {
		t.Fatalf("ChecksumBits = %v, want empty", payload.ChecksumBits)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/analysis/... -run TestAttribute -v`
Expected: FAIL — `Attribute` undefined.

- [ ] **Step 3: Write the implementation**

`backend/internal/application/infrared/analysis/attribution.go`:

```go
package applicationinfraredanalysis

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

type StateAttribution struct {
	StateId    uuid.UUID
	BitOffsets []int
	ValueBits  map[string][]int
}

type AnalysisPayload struct {
	FrameBitLength int
	BaselineBits   []int
	ChecksumBits   []int
	VolatileBits   []int
	States         []StateAttribution
}

// Attribute discovers checksum bits and per-state bit ownership from a
// baseline case's bits plus every non-baseline case's bits, given which
// single state each non-baseline case targets (by this plan's OFAT
// construction, exactly one). See this task's design note in the
// implementation plan for the intersection-based checksum-discovery
// rationale.
func Attribute(
	baselineBits []int,
	caseBits map[uuid.UUID][]int,
	caseTargetState map[uuid.UUID]uuid.UUID,
	caseTargetValue map[uuid.UUID]string,
	volatile map[int]struct{},
) (AnalysisPayload, error) {
	perStateDelta := make(map[uuid.UUID]map[int]struct{})
	for caseId, bits := range caseBits {
		if len(bits) != len(baselineBits) {
			return AnalysisPayload{}, fmt.Errorf("case %s bit length %d does not match baseline length %d", caseId, len(bits), len(baselineBits))
		}
		stateId, ok := caseTargetState[caseId]
		if !ok {
			return AnalysisPayload{}, fmt.Errorf("case %s has no target state", caseId)
		}
		if _, ok := perStateDelta[stateId]; !ok {
			perStateDelta[stateId] = make(map[int]struct{})
		}
		for i := range bits {
			if _, isVolatile := volatile[i]; isVolatile {
				continue
			}
			if bits[i] != baselineBits[i] {
				perStateDelta[stateId][i] = struct{}{}
			}
		}
	}

	checksumBits := make(map[int]struct{})
	stateIds := make([]uuid.UUID, 0, len(perStateDelta))
	for stateId := range perStateDelta {
		stateIds = append(stateIds, stateId)
	}
	for i := 0; i < len(stateIds); i++ {
		for j := i + 1; j < len(stateIds); j++ {
			for bit := range perStateDelta[stateIds[i]] {
				if _, sharedWithOther := perStateDelta[stateIds[j]][bit]; sharedWithOther {
					checksumBits[bit] = struct{}{}
				}
			}
		}
	}

	states := make([]StateAttribution, 0, len(perStateDelta))
	for stateId, delta := range perStateDelta {
		offsets := make([]int, 0, len(delta))
		for bit := range delta {
			if _, isChecksum := checksumBits[bit]; !isChecksum {
				offsets = append(offsets, bit)
			}
		}
		sort.Ints(offsets)

		valueBits := make(map[string][]int)
		for caseId, target := range caseTargetState {
			if target != stateId {
				continue
			}
			bits := caseBits[caseId]
			pattern := make([]int, len(offsets))
			for k, offset := range offsets {
				pattern[k] = bits[offset]
			}
			valueBits[caseTargetValue[caseId]] = pattern
		}

		states = append(states, StateAttribution{StateId: stateId, BitOffsets: offsets, ValueBits: valueBits})
	}
	sort.Slice(states, func(i, j int) bool { return states[i].StateId.String() < states[j].StateId.String() })

	checksumList := make([]int, 0, len(checksumBits))
	for bit := range checksumBits {
		checksumList = append(checksumList, bit)
	}
	sort.Ints(checksumList)

	volatileList := make([]int, 0, len(volatile))
	for bit := range volatile {
		volatileList = append(volatileList, bit)
	}
	sort.Ints(volatileList)

	return AnalysisPayload{
		FrameBitLength: len(baselineBits),
		BaselineBits:   baselineBits,
		ChecksumBits:   checksumList,
		VolatileBits:   volatileList,
		States:         states,
	}, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/analysis/... -v`
Expected: PASS on every test in the package.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/infrared/analysis/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/infrared/analysis/attribution.go backend/internal/application/infrared/analysis/attribution_test.go
git commit -m "feat: add IR checksum discovery and per-state bit attribution"
```

---

### Task 7: `goja` execution engine

**Files:**
- Create: `backend/internal/infrastructure/jsengine/goja.go`
- Create: `backend/internal/infrastructure/jsengine/goja_test.go`

**Interfaces:**
- Produces: `infrastructurejsengine.RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error)`.

`goja` (`github.com/dop251/goja`) is a new dependency — a pure-Go ECMAScript interpreter, no cgo, no subprocess. This lives under `infrastructure/` (not `application/`) because it wraps an external library the same way `infrastructure/llm/claude`/`infrastructure/llm/openai` wrap their SDKs — general-purpose, not IR-specific in principle, even though IR is its first consumer (same framing the LLM `ClientFactory` was given in the merged prior plan).

**Contract:** the generated `EncoderSource` must define a global function `function encode(state)` taking a plain object (state name → string value) and returning an array of integers (mark/space microsecond durations). `RunEncoder` loads the source into a fresh `goja.Runtime` (one per call — no shared/pooled VM state between encoder runs, since a previous run's global mutations must never leak into the next), calls `encode(state)`, and converts the result to `[]int32`, validating it's a non-empty array of numbers. Execution is bounded by `timeout` via `goja`'s documented interrupt mechanism (`vm.Interrupt(...)` called from a `time.AfterFunc` timer) so a generated encoder with an infinite loop cannot hang the calling goroutine forever.

- [ ] **Step 1: Add the dependency**

Run: `cd backend && go get github.com/dop251/goja@latest && go mod tidy`

- [ ] **Step 2: Write the failing tests**

`backend/internal/infrastructure/jsengine/goja_test.go`:

```go
package infrastructurejsengine

import (
	"strings"
	"testing"
	"time"
)

func TestRunEncoderReturnsArrayFromValidSource(t *testing.T) {
	source := `function encode(state) { return state.POWER === "ON" ? [9000, 4500, 560, 560] : [9000, 4500, 560, 1690]; }`
	result, err := RunEncoder(source, map[string]string{"POWER": "ON"}, time.Second)
	if err != nil {
		t.Fatalf("RunEncoder() error = %v, want nil", err)
	}
	want := []int32{9000, 4500, 560, 560}
	if len(result) != len(want) {
		t.Fatalf("RunEncoder() = %v, want %v", result, want)
	}
	for i := range want {
		if result[i] != want[i] {
			t.Fatalf("RunEncoder() = %v, want %v", result, want)
		}
	}
}

func TestRunEncoderPropagatesThrownError(t *testing.T) {
	source := `function encode(state) { throw new Error("unsupported state"); }`
	_, err := RunEncoder(source, map[string]string{}, time.Second)
	if err == nil {
		t.Fatal("RunEncoder() error = nil, want the thrown error propagated")
	}
	if !strings.Contains(err.Error(), "unsupported state") {
		t.Fatalf("RunEncoder() error = %v, want it to mention the thrown message", err)
	}
}

func TestRunEncoderErrorsOnNonArrayReturn(t *testing.T) {
	source := `function encode(state) { return "not an array"; }`
	_, err := RunEncoder(source, map[string]string{}, time.Second)
	if err == nil {
		t.Fatal("RunEncoder() error = nil, want an error for a non-array return value")
	}
}

func TestRunEncoderTimesOutOnInfiniteLoop(t *testing.T) {
	source := `function encode(state) { while (true) {} }`
	start := time.Now()
	_, err := RunEncoder(source, map[string]string{}, 100*time.Millisecond)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("RunEncoder() error = nil, want a timeout error")
	}
	if elapsed > 2*time.Second {
		t.Fatalf("RunEncoder() took %v, want it to return promptly after the timeout", elapsed)
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/infrastructure/jsengine/... -v`
Expected: FAIL — `RunEncoder` undefined.

- [ ] **Step 4: Write the implementation**

`backend/internal/infrastructure/jsengine/goja.go`:

```go
package infrastructurejsengine

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// RunEncoder loads source into a fresh goja VM (never reused across calls),
// calls its global encode(state) function with state, and converts the
// result to a mark/space duration array. Execution is interrupted if it
// runs past timeout, so a generated encoder with a runaway loop cannot
// hang the caller forever.
func RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error) {
	vm := goja.New()

	timer := time.AfterFunc(timeout, func() {
		vm.Interrupt(fmt.Errorf("encoder execution exceeded %s", timeout))
	})
	defer timer.Stop()

	if _, err := vm.RunString(source); err != nil {
		return nil, fmt.Errorf("failed to load encoder source: %w", err)
	}

	encodeFn, ok := goja.AssertFunction(vm.Get("encode"))
	if !ok {
		return nil, fmt.Errorf("encoder source does not define a callable encode function")
	}

	stateValue := vm.ToValue(state)
	result, err := encodeFn(goja.Undefined(), stateValue)
	if err != nil {
		return nil, fmt.Errorf("encoder threw an error: %w", err)
	}

	exported := result.Export()
	items, ok := exported.([]interface{})
	if !ok {
		return nil, fmt.Errorf("encoder returned %T, want an array of numbers", exported)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("encoder returned an empty array")
	}

	durations := make([]int32, len(items))
	for i, item := range items {
		switch v := item.(type) {
		case int64:
			durations[i] = int32(v)
		case float64:
			durations[i] = int32(v)
		default:
			return nil, fmt.Errorf("encoder array element %d is %T, want a number", i, item)
		}
	}
	return durations, nil
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/infrastructure/jsengine/... -v`
Expected: PASS on all four tests.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/infrastructure/jsengine/`
Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add backend/go.mod backend/go.sum backend/internal/infrastructure/jsengine/
git commit -m "feat: add goja-backed encoder execution engine"
```

---

### Task 8: LLM coder generation (`WriteCoder`)

**Files:**
- Create: `backend/internal/application/infrared/coder_generation/llm_coder.go`
- Create: `backend/internal/application/infrared/coder_generation/llm_coder_test.go`

**Interfaces:**
- Consumes: `domaincontractsllm.Client`/`GenerateTextRequest`/`GenerateTextResult` (merged prior plan, same contract Plan B1's `WriteScript` already consumes); `applicationinfraredanalysis.AnalysisPayload` (Task 6); `domainmodels.InfraredState`, `domainmodels.InfraredStateDeviceDefinition`.
- Produces: `applicationinfraredcodergeneration.WriteCoder(ctx context.Context, client domaincontractsllm.Client, deviceBrand string, deviceModel string, states []domainmodels.InfraredState, definitions []domainmodels.InfraredStateDeviceDefinition, payload applicationinfraredanalysis.AnalysisPayload) (Coder, error)`, where `Coder{EncoderSource, DecoderSource, SummaryReadme, DetailReadme string}`.

This plan's second consumer of the merged LLM provider abstraction. Unlike the original design spec's framing ("the encoder/decoder+README generation call... leaves [`ResponseSchema`] nil and expects free-form code/text back"), this task uses `ResponseSchema` for a **structured** JSON response with four string fields — the same mechanism Plan B1's `WriteScript` already uses successfully for its own structured output, and strictly more robust than parsing free-form text for four distinct fields via ad-hoc delimiters. This is a deliberate deviation from the spec's original framing, made because the reconciled tooling makes it strictly better, not because the spec was followed incorrectly.

- [ ] **Step 1: Write the failing test**

`backend/internal/application/infrared/coder_generation/llm_coder_test.go` — uses a hand-written fake `domaincontractsllm.Client`, same convention as Plan B1's `llm_script_test.go`:

```go
package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"testing"

	applicationinfraredanalysis "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/analysis"
	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type fakeLlmClient struct {
	responseText string
	err          error
	lastRequest  domaincontractsllm.GenerateTextRequest
}

func (f *fakeLlmClient) GenerateText(_ context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	f.lastRequest = req
	return domaincontractsllm.GenerateTextResult{Text: f.responseText}, f.err
}

func TestWriteCoderParsesStructuredResponse(t *testing.T) {
	powerId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	definitions := []domainmodels.InfraredStateDeviceDefinition{{InfraredStateId: powerId, Options: []string{"ON", "OFF"}}}
	payload := applicationinfraredanalysis.AnalysisPayload{
		FrameBitLength: 4,
		BaselineBits:   []int{0, 0, 0, 0},
		States:         []applicationinfraredanalysis.StateAttribution{{StateId: powerId, BitOffsets: []int{0}, ValueBits: map[string][]int{"ON": {0}, "OFF": {1}}}},
	}

	responseBody, _ := json.Marshal(map[string]string{
		"encoder_source": "function encode(state) { return [9000, 4500]; }",
		"decoder_source": "function decode(raw) { return {}; }",
		"summary_readme": "Polytron PAC-09HDN infrared protocol.",
		"detail_readme":  "POWER is encoded at bit offset 0.",
	})
	client := &fakeLlmClient{responseText: string(responseBody)}

	coder, err := WriteCoder(context.Background(), client, "Polytron", "PAC-09HDN", states, definitions, payload)
	if err != nil {
		t.Fatalf("WriteCoder() error = %v, want nil", err)
	}
	if coder.EncoderSource == "" || coder.DecoderSource == "" || coder.SummaryReadme == "" || coder.DetailReadme == "" {
		t.Fatalf("WriteCoder() = %+v, want all four fields populated", coder)
	}

	if client.lastRequest.ResponseSchema == nil {
		t.Fatalf("GenerateText() request had no ResponseSchema — structured output was not requested")
	}
	if client.lastRequest.MaxOutputTokens <= 0 {
		t.Fatalf("GenerateText() request had MaxOutputTokens = %d, want > 0", client.lastRequest.MaxOutputTokens)
	}
}

func TestWriteCoderPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: context.DeadlineExceeded}
	_, err := WriteCoder(context.Background(), client, "Polytron", "PAC-09HDN", nil, nil, applicationinfraredanalysis.AnalysisPayload{})
	if err == nil {
		t.Fatal("WriteCoder() error = nil, want propagated error")
	}
}

func TestWriteCoderErrorsOnMalformedResponse(t *testing.T) {
	client := &fakeLlmClient{responseText: "not json"}
	_, err := WriteCoder(context.Background(), client, "Polytron", "PAC-09HDN", nil, nil, applicationinfraredanalysis.AnalysisPayload{})
	if err == nil {
		t.Fatal("WriteCoder() error = nil, want a malformed-response error")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -v`
Expected: FAIL — `WriteCoder` undefined.

- [ ] **Step 3: Write the implementation**

`backend/internal/application/infrared/coder_generation/llm_coder.go`:

```go
package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	applicationinfraredanalysis "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/analysis"
)

const responseSchema = `{
	"type": "object",
	"properties": {
		"encoder_source": {"type": "string"},
		"decoder_source": {"type": "string"},
		"summary_readme": {"type": "string"},
		"detail_readme": {"type": "string"}
	},
	"required": ["encoder_source", "decoder_source", "summary_readme", "detail_readme"],
	"additionalProperties": false
}`

type Coder struct {
	EncoderSource string
	DecoderSource string
	SummaryReadme string
	DetailReadme  string
}

type coderResponse struct {
	EncoderSource string `json:"encoder_source"`
	DecoderSource string `json:"decoder_source"`
	SummaryReadme string `json:"summary_readme"`
	DetailReadme  string `json:"detail_readme"`
}

// WriteCoder asks the LLM to write a JavaScript encoder/decoder pair plus
// two README fields, given the deterministic bit-analysis payload as
// context. The LLM's job is pattern recognition and code generation over
// data Go has already extracted — it never re-derives which bits belong to
// which state; that's payload.States, computed in Task 6.
func WriteCoder(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	states []domainmodels.InfraredState,
	definitions []domainmodels.InfraredStateDeviceDefinition,
	payload applicationinfraredanalysis.AnalysisPayload,
) (Coder, error) {
	stateNameById := make(map[string]string, len(states))
	for _, state := range states {
		stateNameById[state.Id.String()] = state.Name
	}

	var promptBuilder strings.Builder
	fmt.Fprintf(&promptBuilder, "Device: %s %s\n\n", deviceBrand, deviceModel)
	fmt.Fprintf(&promptBuilder, "Frame is %d bits. Baseline bits: %v. Checksum bit offsets (do not treat as state data): %v. Volatile bit offsets (rolling/timestamp, ignore): %v.\n\n",
		payload.FrameBitLength, payload.BaselineBits, payload.ChecksumBits, payload.VolatileBits)
	promptBuilder.WriteString("Per-state bit ownership and observed value patterns:\n")
	for _, s := range payload.States {
		fmt.Fprintf(&promptBuilder, "- %s: bit offsets %v\n", stateNameById[s.StateId.String()], s.BitOffsets)
		for value, bits := range s.ValueBits {
			fmt.Fprintf(&promptBuilder, "  value %q -> bits %v\n", value, bits)
		}
	}
	promptBuilder.WriteString("\nWrite a JavaScript encoder function `function encode(state)` taking an object keyed by state name (e.g. state.POWER, state.MODE) with string values, returning an array of mark/space microsecond durations for the full IR frame including header and any checksum computation your analysis of the bit layout implies. Also write a best-effort `function decode(raw)` inverse (not verified by this pipeline). Then write a short protocol summary and a longer detailed explanation of the encoding (frame structure, checksum algorithm if any, per-state bit meaning). Respond as JSON matching the given schema.")

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at reverse-engineering infrared remote control protocols and writing correct JavaScript encoders/decoders from bit-level analysis data.",
		Prompt:          promptBuilder.String(),
		MaxOutputTokens: 8192,
		ResponseSchema:  []byte(responseSchema),
	})
	if err != nil {
		return Coder{}, domainmodels.NewError("failed to generate encoder/decoder", domainmodels.ErrTypeFailure, err)
	}

	var response coderResponse
	if err := json.Unmarshal([]byte(result.Text), &response); err != nil {
		return Coder{}, domainmodels.NewError("llm returned malformed coder response", domainmodels.ErrTypeFailure, err)
	}

	return Coder{
		EncoderSource: response.EncoderSource,
		DecoderSource: response.DecoderSource,
		SummaryReadme: response.SummaryReadme,
		DetailReadme:  response.DetailReadme,
	}, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -v`
Expected: PASS on all three tests.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/infrared/coder_generation/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/infrared/coder_generation/
git commit -m "feat: add LLM-driven encoder/decoder generation"
```

---

### Task 9: Cursor advancement and the `RECORDING → ANALYZING` trigger

**Files:**
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: `domaincontractsrepository.InfraredStateDeviceRecordCase.CountAcceptedRawByCaseId`/`ListBySessionId` (Plan B1, already exist on the domain contract but with no fake override yet), `domaincontractsrepository.InfraredRecordSession.UpdateCurrentRecordCaseIdById` (Plan B1, already exists — the real fake's `fakeSessionRepository` already tracks it via `currentCaseId`/`CurrentCaseId()`, added during Plan B1's final-review fix — read that fake before writing anything here, it is NOT the simple map some earlier drafts of this task assumed).
- Produces: `AcceptRaw` now advances the session's cursor once a case reaches its required 2 accepted raws, and triggers the `RECORDING → ANALYZING` transition (+ launches Task 10's async job) once no `PENDING` case remains.

**Why this task exists:** Plan B1 built raw capture (`CaptureIrRaw`) and per-raw accept/discard (`AcceptRaw`/`DiscardRaw`), but never built the piece that decides *when a case is done* and *what happens next* — `AcceptRaw` today is a bare one-line call to `UpdateRawStatusById` with no cursor logic at all. This task adds exactly that, since it's a prerequisite for anything in this plan (`ANALYZING` never starts if nothing ever leaves `RECORDING`).

**Read the real fakes in `usecase_test.go` before writing anything** — `fakeSessionRepository.statusUpdates` is a mutex-protected `[]string`, read only via its `StatusUpdates()` snapshot accessor (never the raw field, even from within the same package — that convention exists specifically because a background goroutine may still be writing while a test reads); `fakeSessionRepository.currentCaseId` is a single `*uuid.UUID` (not a per-session map — every test in this file constructs one usecase per session anyway), read via `CurrentCaseId()`. `fakeCaseRepository` currently has `createRawCalls`/`CreateRaw`/`CreateWithStates`/`CreatedCaseIds()`/`UpdateStatusById` (which only records the updated case's *id* into `statusUpdateIds`, not the status value)/`GetById`/`ListRawByCaseId` — there is no `CountAcceptedRawByCaseId` override, no `ListBySessionId` override, and no way yet to look up a raw by id or track *which status* `UpdateStatusById` was called with per case. All of that needs adding in this task, following the file's own established mutex+snapshot-accessor convention for anything a background goroutine might touch concurrently with a test's assertions.

- [ ] **Step 1: Write the failing tests**

Add to `backend/internal/application/infrared/record_session_management/usecase_test.go`:

```go
func (f *fakeCaseRepository) CountAcceptedRawByCaseId(_ context.Context, caseId uuid.UUID) (int, error) {
	return f.acceptedRawCounts[caseId], nil
}
func (f *fakeCaseRepository) ListBySessionId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordCase, error) {
	return f.listBySessionIdResult, nil
}
func (f *fakeCaseRepository) GetRawById(_ context.Context, rawId uuid.UUID) (*domainmodels.InfraredStateDeviceRecordRaw, error) {
	return f.rawById[rawId], nil
}

// StatusUpdatesForCase returns a snapshot of every status this fake's
// UpdateStatusById has been called with for the given case, in call order —
// mirrors the file's existing StatusUpdates()/CurrentCaseId() convention
// for state a background goroutine might mutate concurrently with a test.
func (f *fakeCaseRepository) StatusUpdatesForCase(caseId uuid.UUID) []domainmodels.InfraredRecordCaseStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]domainmodels.InfraredRecordCaseStatus(nil), f.statusValuesByCase[caseId]...)
}

func TestAcceptRawAdvancesCursorToNextPendingCase(t *testing.T) {
	sessionId := uuid.New()
	firstCaseId, secondCaseId := uuid.New(), uuid.New()
	rawId := uuid.New()

	sessionRepo := &fakeSessionRepository{}
	caseRepo := &fakeCaseRepository{
		acceptedRawCounts: map[uuid.UUID]int{firstCaseId: 2},
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: firstCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusActive},
			{Id: secondCaseId, InfraredRecordSessionId: sessionId, Step: 2, Status: domainmodels.InfraredRecordCaseStatusPending},
		},
		rawById: map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw{
			rawId: {Id: rawId, InfraredStateDeviceRecordCaseId: firstCaseId},
		},
	}
	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &noopLogger{},
	)

	if err := usecase.AcceptRaw(context.Background(), rawId); err != nil {
		t.Fatalf("AcceptRaw() error = %v, want nil", err)
	}

	firstUpdates := caseRepo.StatusUpdatesForCase(firstCaseId)
	if len(firstUpdates) == 0 || firstUpdates[len(firstUpdates)-1] != domainmodels.InfraredRecordCaseStatusAccepted {
		t.Fatalf("first case status updates = %v, want to end with ACCEPTED", firstUpdates)
	}
	secondUpdates := caseRepo.StatusUpdatesForCase(secondCaseId)
	if len(secondUpdates) == 0 || secondUpdates[len(secondUpdates)-1] != domainmodels.InfraredRecordCaseStatusActive {
		t.Fatalf("second case status updates = %v, want to end with ACTIVE", secondUpdates)
	}
	if got := sessionRepo.CurrentCaseId(); got == nil || *got != secondCaseId {
		t.Fatalf("session current case = %v, want %v", got, secondCaseId)
	}
}

func TestAcceptRawTriggersAnalyzingWhenNoPendingCaseRemains(t *testing.T) {
	sessionId := uuid.New()
	onlyCaseId := uuid.New()
	rawId := uuid.New()

	sessionRepo := &fakeSessionRepository{}
	caseRepo := &fakeCaseRepository{
		acceptedRawCounts: map[uuid.UUID]int{onlyCaseId: 2},
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: onlyCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusActive},
		},
		rawById: map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw{
			rawId: {Id: rawId, InfraredStateDeviceRecordCaseId: onlyCaseId, InfraredRecordSessionId: sessionId},
		},
	}
	// runAnalysisAndGeneration is still Task 9's no-op stub at this point in
	// the plan (Task 10 fills it in), so llmFactory is never actually
	// called by the goroutine AcceptRaw launches — passed anyway for
	// forward-compatibility with Task 10's real body.
	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{err: errors.New("llm not configured for this test")}, &fakeNodeRepository{}, &noopLogger{},
	)

	if err := usecase.AcceptRaw(context.Background(), rawId); err != nil {
		t.Fatalf("AcceptRaw() error = %v, want nil", err)
	}

	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateAnalyzing {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include ANALYZING", sessionRepo.StatusUpdates())
	}
}

func TestAcceptRawDoesNotAdvanceCursorBeforeSecondRawAccepted(t *testing.T) {
	sessionId := uuid.New()
	caseId := uuid.New()
	rawId := uuid.New()

	sessionRepo := &fakeSessionRepository{}
	caseRepo := &fakeCaseRepository{
		acceptedRawCounts: map[uuid.UUID]int{caseId: 1},
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: caseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusActive},
		},
		rawById: map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw{
			rawId: {Id: rawId, InfraredStateDeviceRecordCaseId: caseId},
		},
	}
	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &noopLogger{},
	)

	if err := usecase.AcceptRaw(context.Background(), rawId); err != nil {
		t.Fatalf("AcceptRaw() error = %v, want nil", err)
	}
	if updates := caseRepo.StatusUpdatesForCase(caseId); len(updates) != 0 {
		t.Fatalf("case status updates = %v, want none (only 1 of 2 required raws accepted)", updates)
	}
	if got := sessionRepo.CurrentCaseId(); got != nil {
		t.Fatalf("session current case = %v, want nil (no change before the case's second raw is accepted)", got)
	}
}
```

Extend `fakeCaseRepository`'s struct definition (find it near the top of `usecase_test.go`, alongside its existing `mu sync.Mutex` field) with the new fields these tests need: `acceptedRawCounts map[uuid.UUID]int`, `listBySessionIdResult []domainmodels.InfraredStateDeviceRecordCase`, `rawById map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw`, `statusValuesByCase map[uuid.UUID][]domainmodels.InfraredRecordCaseStatus`. Extend the EXISTING `UpdateStatusById` method (do not replace it — it already appends to `statusUpdateIds`, which an existing `RetryCase` test asserts on) to ALSO append the status value to `statusValuesByCase[id]` under the same `f.mu` lock it already takes. `fakeSessionRepository` needs no changes — `CurrentCaseId()`/`StatusUpdates()` already exist and already do exactly what these tests need.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run TestAcceptRaw -v`
Expected: FAIL — either compile errors (missing fake fields/methods) or behavioral failures, since `AcceptRaw` doesn't yet do any of this.

- [ ] **Step 3: Write the implementation**

**Design note:** `AcceptRaw` is currently called with only a `rawId` — to find which case (and session) that raw belongs to, add a `GetRawById(ctx, rawId uuid.UUID) (*domainmodels.InfraredStateDeviceRecordRaw, error)` method to `domaincontractsrepository.InfraredStateDeviceRecordCase` (the raw's `InfraredStateDeviceRecordCaseId` field, already on the model since Plan B1, gives the case; the case's `InfraredRecordSessionId` field, also already on the model, gives the session — no new repository beyond one new method is needed). Implement it in `infrastructure/repository/infrared_state_device_record_case/postgres.go` following the exact `GetById`-style pattern already in that file.

Rewrite `AcceptRaw` in `usecase.go`:

```go
func (u *usecase) AcceptRaw(ctx context.Context, rawId uuid.UUID) error {
	const tag = "infrared/record_session_management/AcceptRaw"

	if err := u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusAccepted, nil); err != nil {
		return err
	}

	raw, err := u.recordCase.GetRawById(ctx, rawId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up accepted raw", domainmodels.LoggerMeta{"err": err, "raw_id": rawId})
		return err
	}

	acceptedCount, err := u.recordCase.CountAcceptedRawByCaseId(ctx, raw.InfraredStateDeviceRecordCaseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to count accepted raws", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	if acceptedCount < 2 {
		return nil
	}

	if err := u.recordCase.UpdateStatusById(ctx, raw.InfraredStateDeviceRecordCaseId, domainmodels.InfraredRecordCaseStatusAccepted); err != nil {
		u.logger.Error(ctx, tag, "failed to mark case accepted", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}

	cases, err := u.recordCase.ListBySessionId(ctx, raw.InfraredRecordSessionId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list cases for cursor advancement", domainmodels.LoggerMeta{"err": err, "session_id": raw.InfraredRecordSessionId})
		return err
	}

	var next *domainmodels.InfraredStateDeviceRecordCase
	for i := range cases {
		if cases[i].Status != domainmodels.InfraredRecordCaseStatusPending {
			continue
		}
		if next == nil || cases[i].Step < next.Step {
			next = &cases[i]
		}
	}

	if next != nil {
		if err := u.recordCase.UpdateStatusById(ctx, next.Id, domainmodels.InfraredRecordCaseStatusActive); err != nil {
			u.logger.Error(ctx, tag, "failed to activate next case", domainmodels.LoggerMeta{"err": err, "case_id": next.Id})
			return err
		}
		if err := u.session.UpdateCurrentRecordCaseIdById(ctx, raw.InfraredRecordSessionId, &next.Id); err != nil {
			u.logger.Error(ctx, tag, "failed to advance session cursor", domainmodels.LoggerMeta{"err": err, "session_id": raw.InfraredRecordSessionId})
			return err
		}
		u.broadcastBestEffort(ctx, tag, raw.InfraredRecordSessionId, domainmodels.InfraredRecordingStateRecording, &next.Id)
		return nil
	}

	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, raw.InfraredRecordSessionId, nil); err != nil {
		u.logger.Error(ctx, tag, "failed to clear session cursor before analysis", domainmodels.LoggerMeta{"err": err, "session_id": raw.InfraredRecordSessionId})
		return err
	}
	u.transition(ctx, tag, raw.InfraredRecordSessionId, domainmodels.InfraredRecordingStateAnalyzing)

	go u.runAnalysisAndGeneration(raw.InfraredRecordSessionId)

	return nil
}

func (u *usecase) broadcastBestEffort(ctx context.Context, tag string, sessionId uuid.UUID, state string, currentCaseId *uuid.UUID) {
	if err := u.broadcaster.Send(ctx, domainmodels.InfraredRecordSessionEvent{SessionId: sessionId, RecordingState: state, CurrentRecordCaseId: currentCaseId}); err != nil {
		u.logger.Warn(ctx, tag, "failed to broadcast", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
	}
}
```

`runAnalysisAndGeneration` is Task 10's job — this task only needs to compile against a placeholder method with this exact name and a `sessionId uuid.UUID` parameter; Task 10 fills in its body. Add a minimal stub here (`func (u *usecase) runAnalysisAndGeneration(sessionId uuid.UUID) {}`) so this task's tests pass and compile in isolation — Task 10 replaces the stub body, not the method's existence or signature, so this is not a placeholder that ships (the plan's later task closes it before the branch is done, exactly like Plan B1's Task 13/15 split for `CaptureIrRaw` — except here the split is between consecutive tasks in the SAME plan, both of which land before final review, so there is no risk of a stub shipping unnoticed).

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -v`
Expected: PASS on every test in the package, including every test from Plan B1's Task 13 (must still pass unmodified) and the new ones above.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/infrared_state_device_record_case/ internal/application/infrared/record_session_management/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/contracts/repository/infrared_state_device_record_case.go \
        backend/internal/infrastructure/repository/infrared_state_device_record_case/ \
        backend/internal/application/infrared/record_session_management/
git commit -m "feat: add cursor advancement and RECORDING to ANALYZING trigger"
```

---

### Task 10: Analysis + generation async job

**Files:**
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: every repository this usecase already holds (Plan B1), `applicationinfraredanalysis.SegmentFrames`/`DemodulateBits`/`DetectVolatileBits`/`UnionVolatileBits`/`Attribute` (Tasks 4-6), `applicationinfraredcodergeneration.WriteCoder` (Task 8), `domaincontractsrepository.InfraredStateCoder` (Task 3), a new `EncoderRunner` interface (declared locally in this usecase's package, same pattern as `LlmClientFactory` from Plan B1's Task 13) wrapping Task 7's `infrastructurejsengine.RunEncoder`.
- Produces: `runAnalysisAndGeneration`'s real body (replacing Task 9's stub), plus a new `GetCoderBySessionId` method on `RecordSessionManagement`.

This is this plan's second background-goroutine job, structurally identical to Plan B1's `runCaseGeneration`: detached `context.Background()` + its own timeout, logs from inside the goroutine (nothing else can observe an error there), transitions the session on completion or failure.

- [ ] **Step 1: Add `EncoderRunner` and `GetCoderBySessionId` to the interfaces**

Add to `backend/internal/domain/usecases/infrared/record_session_management.go`:

```go
GetCoderBySessionId(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredStateCoder, error)
```

(as a new method on the existing `RecordSessionManagement` interface).

In `usecase.go`, add:

```go
// EncoderRunner is the subset of infrastructurejsengine's exported surface
// this usecase needs — declared locally so the usecase doesn't import the
// infrastructure package directly, same pattern as LlmClientFactory.
type EncoderRunner interface {
	RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error)
}
```

Add `encoderRunner EncoderRunner`, `coder domaincontractsrepository.InfraredStateCoder`, and `state domaincontractsrepository.InfraredState` (if not already present — it already is, from Plan B1) fields to the `usecase` struct and `NewUsecaseImpl`'s parameter list, appended after the existing `node` parameter and before `logger`. Update every existing `NewUsecaseImpl(...)` call site in `usecase_test.go` (there are several from Plan B1) to pass a new `&fakeEncoderRunner{}` and `&fakeCoderRepository{}` at that position — check each call site and add the two new arguments in the right place; do not reorder any existing argument.

- [ ] **Step 2: Write the failing tests**

Add to `usecase_test.go`:

```go
type fakeEncoderRunner struct {
	err       error
	responses map[string][]int32
}

func (f *fakeEncoderRunner) RunEncoder(source string, state map[string]string, _ time.Duration) ([]int32, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []int32{9000, 4500}, nil
}

type fakeCoderRepository struct {
	domaincontractsrepository.InfraredStateCoder
	created domainmodels.InfraredStateCoder
	getResult *domainmodels.InfraredStateCoder
}

func (f *fakeCoderRepository) Create(_ context.Context, coder domainmodels.InfraredStateCoder) (uuid.UUID, error) {
	f.created = coder
	return uuid.New(), nil
}
func (f *fakeCoderRepository) GetBySessionId(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	return f.getResult, nil
}

func TestRunAnalysisAndGenerationPersistsCoderAndStaysInFunctionGenerating(t *testing.T) {
	sessionId := uuid.New()
	powerId := uuid.New()

	sessionRepo := &fakeSessionRepository{}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: uuid.New(), InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
	}
	coderRepo := &fakeCoderRepository{}
	llmFactory := &fakeLlmClientFactory{}
	encoderRunner := &fakeEncoderRunner{}

	// NewUsecaseImpl returns the domainusecasesinfrared.RecordSessionManagement
	// interface, but runAnalysisAndGeneration is unexported — assert back to
	// the concrete *usecase type to call it directly. Name this "impl", not
	// "usecase": a local variable named "usecase" would shadow the package's
	// own "usecase" struct type, making a later `.(*usecase)` assertion a
	// compile error (the identifier would resolve to the variable, not the
	// type, in that scope).
	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if coderRepo.created.InfraredRecordSessionId != sessionId {
		t.Fatalf("created coder session id = %v, want %v", coderRepo.created.InfraredRecordSessionId, sessionId)
	}
	if coderRepo.created.EncoderSource == "" {
		t.Fatal("created coder has empty EncoderSource")
	}

	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateFunctionGenerating {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include FUNCTION_GENERATING", sessionRepo.StatusUpdates())
	}
}

func TestRunAnalysisAndGenerationFailsSessionWhenEncoderThrows(t *testing.T) {
	sessionId := uuid.New()
	sessionRepo := &fakeSessionRepository{}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: uuid.New(), InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
	}
	coderRepo := &fakeCoderRepository{}
	encoderRunner := &fakeEncoderRunner{err: errors.New("encoder threw")}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, encoderRunner, coderRepo, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if coderRepo.created.Id != uuid.Nil {
		t.Fatal("coder was created despite the encoder smoke test failing, want no persisted row")
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateFailed {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include FAILED", sessionRepo.StatusUpdates())
	}
}

func TestGetCoderBySessionIdDelegatesToRepository(t *testing.T) {
	sessionId := uuid.New()
	want := &domainmodels.InfraredStateCoder{Id: uuid.New(), InfraredRecordSessionId: sessionId}
	coderRepo := &fakeCoderRepository{getResult: want}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, &noopLogger{},
	)

	got, err := usecase.GetCoderBySessionId(context.Background(), sessionId)
	if err != nil || got != want {
		t.Fatalf("GetCoderBySessionId() = %v, %v, want %v, nil", got, err, want)
	}
}
```

You will need to add `listByDeviceTypeIdResult []domainmodels.InfraredState` to `fakeStateRepository`'s struct definition and change its existing `ListByDeviceTypeId` method (which currently always returns `nil, nil`) to return that field instead, and confirm `fakeCaseRepository.ListRawByCaseId` (which also currently always returns `nil, nil`) returning no raws for the one fixture case above is fine for these two tests — check exactly what data `runAnalysisAndGeneration` needs (Step 4 below, where the real implementation is written) and make sure the fakes in these two tests supply it. A case with only one case (the baseline, `Step: 1`, no accepted raws in the fake) is deliberately the simplest input that exercises the full pipeline without needing to hand-construct a realistic multi-case OFAT sweep with real raw captures in a test fixture: with zero accepted raws, `buildAnalysisPayload` never sets a baseline and `Attribute` receives empty maps throughout, still returning a valid (empty-`States`) payload with no error — a legitimate degenerate input for this test's purposes (checking persistence/transition wiring, not analysis correctness — that's Tasks 4-6's job, already covered there with real bit patterns).

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run "TestRunAnalysisAndGeneration|TestGetCoderBySessionId" -v`
Expected: FAIL — compile errors from the new constructor parameters and missing real implementation.

- [ ] **Step 4: Write the implementation**

Replace Task 9's stub with the real body:

```go
const analysisTimeout = 3 * time.Minute
const encoderSmokeTestTimeout = 2 * time.Second

func (u *usecase) runAnalysisAndGeneration(sessionId uuid.UUID) {
	const tag = "infrared/record_session_management/runAnalysisAndGeneration"

	ctx, cancel := context.WithTimeout(context.Background(), analysisTimeout)
	defer cancel()

	session, err := u.session.GetById(ctx, sessionId)
	if err != nil || session == nil {
		u.logger.Error(ctx, tag, "failed to look up session", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	device, err := u.device.GetById(ctx, session.InfraredDeviceId)
	if err != nil || device == nil {
		u.logger.Error(ctx, tag, "failed to look up device", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	states, err := u.state.ListByDeviceTypeId(ctx, device.InfraredDeviceTypeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list states", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	definitions, err := u.definition.ListByDeviceId(ctx, session.InfraredDeviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list definitions", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	cases, err := u.recordCase.ListBySessionId(ctx, sessionId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list cases", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	payload, baselineState, err := u.buildAnalysisPayload(ctx, tag, sessionId, cases)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to analyze recorded cases", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFunctionGenerating)

	client, err := u.llmFactory.Current(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	written, err := applicationinfraredcodergeneration.WriteCoder(ctx, client, device.Brand, device.Model, states, definitions, payload)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to write coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	if _, err := u.encoderRunner.RunEncoder(written.EncoderSource, baselineState, encoderSmokeTestTimeout); err != nil {
		u.logger.Error(ctx, tag, "generated encoder failed its smoke test", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	if _, err := u.coder.Create(ctx, domainmodels.InfraredStateCoder{
		InfraredDeviceId:        session.InfraredDeviceId,
		InfraredRecordSessionId: sessionId,
		EncoderSource:           written.EncoderSource,
		DecoderSource:           written.DecoderSource,
		SummaryReadme:           written.SummaryReadme,
		DetailReadme:            written.DetailReadme,
	}); err != nil {
		u.logger.Error(ctx, tag, "failed to persist coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
}

// buildAnalysisPayload turns every case's accepted raws into the pure
// analysis package's inputs and runs the full frame -> demodulate ->
// volatile -> attribute pipeline. Returns the payload plus the baseline
// case's target state (name -> value) for the encoder smoke test.
func (u *usecase) buildAnalysisPayload(ctx context.Context, tag string, sessionId uuid.UUID, cases []domainmodels.InfraredStateDeviceRecordCase) (applicationinfraredanalysis.AnalysisPayload, map[string]string, error) {
	var baselineBits []int
	baselineState := make(map[string]string)
	caseBits := make(map[uuid.UUID][]int)
	caseTargetState := make(map[uuid.UUID]uuid.UUID)
	caseTargetValue := make(map[uuid.UUID]string)
	var volatileSets []map[int]struct{}

	for _, c := range cases {
		raws, err := u.recordCase.ListRawByCaseId(ctx, c.Id)
		if err != nil {
			return applicationinfraredanalysis.AnalysisPayload{}, nil, err
		}
		var accepted [][]int32
		for _, r := range raws {
			if r.Status == domainmodels.InfraredRecordRawStatusAccepted {
				var durations []int32
				if err := json.Unmarshal(r.RawData, &durations); err != nil {
					return applicationinfraredanalysis.AnalysisPayload{}, nil, err
				}
				accepted = append(accepted, durations)
			}
		}
		if len(accepted) < 2 {
			continue
		}

		frameBitsPerRaw := make([][]int, 0, len(accepted))
		for _, raw := range accepted {
			frames := applicationinfraredanalysis.SegmentFrames(raw)
			if len(frames) == 0 {
				continue
			}
			frameBitsPerRaw = append(frameBitsPerRaw, applicationinfraredanalysis.DemodulateBits(frames[0]))
		}
		if len(frameBitsPerRaw) < 2 {
			continue
		}

		volatile, err := applicationinfraredanalysis.DetectVolatileBits(frameBitsPerRaw[0], frameBitsPerRaw[1])
		if err != nil {
			return applicationinfraredanalysis.AnalysisPayload{}, nil, err
		}
		volatileSets = append(volatileSets, volatile)

		states, err := u.recordCase.ListStatesByCaseId(ctx, c.Id)
		if err != nil {
			return applicationinfraredanalysis.AnalysisPayload{}, nil, err
		}

		if c.Step == 1 {
			baselineBits = frameBitsPerRaw[0]
			for _, s := range states {
				baselineState[s.InfraredStateId.String()] = s.StateValue
			}
			continue
		}

		caseBits[c.Id] = frameBitsPerRaw[0]
		// OFAT construction (Plan B1, Task 8): a non-baseline case differs
		// from baseline in exactly one state — find it by comparing against
		// the baseline's own recorded state values.
		for _, s := range states {
			if baselineValue, ok := baselineState[s.InfraredStateId.String()]; ok && baselineValue != s.StateValue {
				caseTargetState[c.Id] = s.InfraredStateId
				caseTargetValue[c.Id] = s.StateValue
				break
			}
		}
	}

	volatile := applicationinfraredanalysis.UnionVolatileBits(volatileSets...)
	payload, err := applicationinfraredanalysis.Attribute(baselineBits, caseBits, caseTargetState, caseTargetValue, volatile)
	if err != nil {
		return applicationinfraredanalysis.AnalysisPayload{}, nil, err
	}
	return payload, baselineState, nil
}

func (u *usecase) GetCoderBySessionId(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	return u.coder.GetBySessionId(ctx, sessionId)
}
```

Add `"encoding/json"` and `applicationinfraredcodergeneration "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/coder_generation"` and `applicationinfraredanalysis "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/analysis"` to `usecase.go`'s imports if not already present.

**Note on `baselineState`'s keying:** it's keyed by `InfraredStateId.String()` here (matching `caseTargetState`), but `RunEncoder`/`WriteCoder`'s prompt both expect a STATE NAME key (`state.POWER`), not a UUID. This task's `runAnalysisAndGeneration` passes `baselineState` (UUID-keyed) directly to `u.encoderRunner.RunEncoder` — **this is a real mismatch to fix during implementation**, not something to silently paper over: convert `baselineState` to a name-keyed map (using the already-fetched `states []domainmodels.InfraredState` to map `InfraredStateId → Name`) before calling `RunEncoder`, exactly mirroring how `buildAnalysisPayload`'s own internal maps are UUID-keyed for precision but the LLM-facing and goja-facing surfaces need names. Write a small local helper (`stateIdToName(states) map[string]string`) and use it at the `RunEncoder` call site.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -v`
Expected: PASS on every test in the package, including every test from Plan B1's Task 13 and this plan's Task 9.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... 2>&1 | grep -v "composition/main"` (composition wiring is Task 12; an isolated failure there is expected) `&& go vet ./internal/domain/usecases/infrared/... ./internal/application/infrared/record_session_management/... && gofmt -l internal/domain/usecases/infrared/ internal/application/infrared/record_session_management/`
Expected: no output beyond the expected composition/main isolation.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/usecases/infrared/record_session_management.go \
        backend/internal/application/infrared/record_session_management/
git commit -m "feat: add analysis and encoder/decoder generation async job"
```

---

### Task 11: HTTP endpoint for reading the generated coder

**Files:**
- Modify: `backend/internal/presentation/http/response/infrared.go`
- Modify: `backend/internal/presentation/http/handler/infrared/handler.go`
- Modify: `backend/internal/presentation/http/route/route.go`

**Interfaces:**
- Consumes: `domainusecasesinfrared.RecordSessionManagement.GetCoderBySessionId` (Task 10).
- Produces: `GET /v1/infrared/record-sessions/:id/coder`, gated by the existing `infrared_record_session:get` permission (no new permission needed — reading a session's generated coder is the same read-scope as reading the session itself).

- [ ] **Step 1: Add the response DTO**

Append to `backend/internal/presentation/http/response/infrared.go`:

```go
type InfraredStateCoderResponse struct {
	Id            string `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	EncoderSource string `json:"encoder_source"`
	DecoderSource string `json:"decoder_source"`
	SummaryReadme string `json:"summary_readme"`
	DetailReadme  string `json:"detail_readme"`
	Status        string `json:"status" example:"UNVERIFIED"`
}

func InfraredStateCoder(model domainmodels.InfraredStateCoder) InfraredStateCoderResponse {
	return InfraredStateCoderResponse{
		Id:            UUIDString(model.Id),
		EncoderSource: model.EncoderSource,
		DecoderSource: model.DecoderSource,
		SummaryReadme: model.SummaryReadme,
		DetailReadme:  model.DetailReadme,
		Status:        string(model.Status),
	}
}
```

- [ ] **Step 2: Add the handler method**

Append to `backend/internal/presentation/http/handler/infrared/handler.go`, following `RecordSessionGetById`'s exact shape:

```go
// RecordSessionCoderGetById godoc
//
// @Summary Infrared Record Session Coder Get By ID
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.InfraredStateCoderResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions/{id}/coder [get]
func (h *handler) RecordSessionCoderGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	coder, err := h.recordSessionUseCase.GetCoderBySessionId(c.Request().Context(), id)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if coder == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("coder"))
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredStateCoder(*coder))
}
```

- [ ] **Step 3: Register the route**

Add `RecordSessionCoderGetById(c *echo.Context) error` to the `InfraredHandler` interface in `route.go`, and add to `routeInfrared`:

```go
v1.GET("/infrared/record-sessions/:id/coder", handler.RecordSessionCoderGetById, permission("infrared_record_session:get"))
```

- [ ] **Step 4: Verify**

Run: `cd backend && go build ./... 2>&1 | grep -v "composition/main"` (composition wiring is Task 12) `&& go vet ./internal/presentation/... && gofmt -l internal/presentation/http/`
Expected: no output beyond the expected composition/main isolation.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/presentation/http/response/infrared.go \
        backend/internal/presentation/http/handler/infrared/handler.go \
        backend/internal/presentation/http/route/route.go
git commit -m "feat: add GET endpoint for a session's generated coder"
```

---

### Task 12: Composition wiring

**Files:**
- Modify: `backend/internal/composition/main/infrastructure.go`
- Modify: `backend/internal/composition/main/application.go`

**Interfaces:**
- Consumes: everything from Tasks 3, 7, 10.

Pure wiring — no business logic. This is the task that resolves the isolated `composition/main` build failures Tasks 10-11 leave behind.

- [ ] **Step 1: Wire the new repository and encoder runner**

In `infrastructure.go`, construct `infraredStateCoderRepository := infrastructurerepositoryinfraredstatecoder.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)`, following the exact one-field-plus-one-constructor-line pattern used for every existing entity.

Since Task 10's `EncoderRunner` interface is a single free function (`infrastructurejsengine.RunEncoder`), not a struct with dependencies, wrap it minimally so it satisfies the usecase's interface — the simplest approach is an adapter struct in `infrastructure.go` itself (or a tiny package-level type):

```go
type encoderRunnerAdapter struct{}

func (encoderRunnerAdapter) RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error) {
	return infrastructurejsengine.RunEncoder(source, state, timeout)
}
```

Construct `encoderRunner := encoderRunnerAdapter{}` alongside the other infrastructure fields.

- [ ] **Step 2: Wire the usecase**

In `application.go`, update the existing `applicationinfraredrecordsessionmanagement.NewUsecaseImpl(...)` call site (from Plan B1's Task 15) to add the two new arguments (`l.infra.encoderRunner`, `l.infra.infraredStateCoderRepository`) in the exact position Task 10's constructor signature places them — after `l.infra.nodeRepository`, before `l.infra.logger`.

- [ ] **Step 3: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l . && go test -count=1 ./...`
Expected: no build/vet/gofmt output; every test passes, including every test from this plan and from Plan B1 — this is the task that makes the whole repo build clean again.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/composition/main/
git commit -m "feat: wire IR analysis and encoder/decoder generation into composition root"
```

---

## Self-review notes

- **Spec coverage:** every element of the design spec's B2-scoped pipeline (RECORDING completion → deterministic bit analysis → LLM encoder/decoder generation → goja smoke test → persisted UNVERIFIED coder) has a task. `TEST_CASES_GENERATING`/`TESTING`/hardware verification/MQTT transmit are explicitly out of scope and named as such in the Goal and Global Constraints.
- **Placeholder scan:** the one intentional stub (`runAnalysisAndGeneration`'s empty body, introduced in Task 9) is filled in by the very next task (Task 10) within the same plan, before final review — not a shipped placeholder. No other "TBD"/"implement later" language appears in any task's steps.
- **Type consistency:** `applicationinfraredanalysis.AnalysisPayload`/`StateAttribution` (Task 6) are consumed with identical field names by `WriteCoder` (Task 8) and by the usecase's `buildAnalysisPayload` (Task 10). `domainmodels.InfraredStateCoder`'s field names (Task 1) match the repository's `Create`/`GetBySessionId` (Task 3), the usecase's `Create` call (Task 10), and the response mapper (Task 11) identically. `EncoderRunner`'s signature (Task 10) matches `infrastructurejsengine.RunEncoder`'s real signature (Task 7) and the composition-root adapter (Task 12) exactly.
- **A real design gap, found and resolved during planning, not left implicit:** the original design spec's lifecycle description assumes something advances the session past `RECORDING` and enforces "2 accepted raws per case" before a case is done — but Plan B1 never built that; its `AcceptRaw` was a one-line status update with no cursor logic at all. Task 9 exists specifically to close this gap (the same way Plan B1's own Task 4 closed an analogous "nothing persists the seed data" gap) — without it, nothing in Tasks 10-12 would ever run in production, since `RECORDING → ANALYZING` would never fire.
- **A UUID-vs-name keying mismatch, flagged inline rather than silently resolved:** Task 10's design note calls out explicitly that `buildAnalysisPayload`'s internal maps are UUID-keyed (for precision against the domain model) but `RunEncoder`/`WriteCoder`'s external surfaces need state-name keys (matching how staff and the LLM think about the protocol) — the fix (a `stateIdToName` conversion at the `RunEncoder` call site) is named as something the implementer must do, not something already solved in the given code, because getting this wrong would compile fine but silently break the encoder smoke test's input.
- **Deliberate deviation from the original design spec, with rationale:** the spec's `FUNCTION_GENERATING` step description says the LLM call "leaves [`ResponseSchema`] nil and expects free-form code/text back." Task 8 uses a structured `ResponseSchema` instead (four string fields), because Plan B1's `WriteScript` already proved this mechanism works reliably for extracting multiple distinct fields from one LLM response, and it's strictly more robust than inventing a delimiter-based parsing scheme for exactly the same problem. Noted explicitly rather than silently diverging from the spec's stated approach.
