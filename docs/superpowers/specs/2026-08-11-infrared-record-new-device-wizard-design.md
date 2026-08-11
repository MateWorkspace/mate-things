# Infrared "Record New Device" Wizard (Spec B)

## Context

Spec A (`docs/superpowers/specs/2026-08-11-infrared-record-sessions-list-detail-design.md`)
covers browsing and inspecting existing record sessions. This spec covers
*creating* one: a guided, resumable, WebSocket-driven wizard that walks a
user through defining a new device, generating a recording plan via LLM,
physically capturing IR signals from a real remote, watching automatic
analysis/encoder-generation (including the repair loop built earlier this
project), verifying the result against real hardware, and — if needed —
looping back through re-recording/re-analysis, before finishing.

The backend's record-session lifecycle (`RecordSessionManagement`) already
implements every state transition this flow needs, including a real-time
WebSocket broadcast endpoint
(`GET /infrared/record-sessions/{id}/broadcast`). Investigation during
brainstorming found **no backend endpoint gaps** for this feature — the
wizard is entirely a frontend build on top of what exists, with one
frontend-only server action needed to bridge the WebSocket's auth
requirement.

## 1. Routing & resumability

- **`/apps/infrared/record/new`** — a standalone multi-step *form*
  (wizard steps 1–2 below). No session exists yet, so this route is not
  resumable by design: nothing is created until the form is submitted.
- **`/apps/infrared/record/[id]`** — shared with Spec A. Renders purely
  as a function of `session.recording_state`:
  - `COMPLETED` or `FAILED` → Spec A's static detail view.
  - any other state → the live wizard (steps 3–10 below).
- The wizard state is **never stored client-side** — every render derives
  its current step from the session's real `recording_state` (plus
  whether any cases/raws exist yet), fetched fresh on load and refreshed
  on every WebSocket event. Closing the tab and reopening the URL always
  lands on the correct step. This also means Spec A's detail view and
  Spec B's wizard share the exact same case-list/coder/test-case display
  components — the wizard just layers live-action controls (accept/
  discard, transmit/confirm) on top when the relevant step is active.

## 1a. Restoring the list page's header button

Spec A's list page (`record/page.tsx`) deliberately omits the "Record New
Device" `PageHeader` action until this route exists, to avoid linking to
a 404. This spec's implementation must add it back — a `Link` to
`/apps/infrared/record/new`, gated on the same `infrared_record_session:add`
permission `Start()`'s route already requires.

## 2. Steps 1–2: the pre-session form (`/record/new`)

A two-step form component (`_components/NewRecordSessionForm.tsx`,
client component, local step state only — nothing persisted until
submit).

**Step 1 — Device details:**
- Node select, populated via the existing `listNodes()` — the physical
  IR-capable hardware that will perform the capture. No IR-capability
  filtering (not reliably derivable from existing node/class data); if
  the wrong node is picked, no captures will arrive in step 5, which is
  self-evident to the user at that point.
- Device type select, populated via `listInfraredDeviceTypes()`, with an
  inline "Add new device type" action reusing the existing
  `createInfraredDeviceType()` — added to the select's options
  immediately on success, no page reload.
- Brand, Model text fields (free text, become
  `InfraredDevice.Brand`/`Model`).

**Step 2 — State values:**
- For the selected device type's states (`listInfraredStates()`, with an
  inline "Add new state" action reusing `createInfraredState()`), one
  input per state:
  - `ENUM` states: an add/remove list of string option values.
  - `RANGE` states: min, max, step numeric inputs.
- These become `StartRecordSessionRequest.Definitions` — the concrete,
  per-device legal values Settings' state definitions (name + type only)
  don't capture.

Submitting calls the existing `Start()`/`POST /infrared/record-sessions`
via a new server action (`_lib/actions.ts`,
`startRecordSessionAction`), which redirects to `/record/{id}` on
success. No waiting for case generation — that runs server-side and the
wizard picks it up via the state-derived rendering above.

## 3. Steps 3–10: the live wizard (`/record/[id]`)

A `RecordSessionWizard` client component renders a 10-step indicator
(current step highlighted, completed steps checked) and the active
step's UI, both derived from session/case/coder/test-case data refetched
on every WebSocket event (see §4).

| # | Step | Derived from | UI |
|---|---|---|---|
| 3 | Case building by AI | `recording_state == CASES_GENERATING` | Waiting indicator — "Generating a recording plan..." |
| 3.5 | Case plan review (wizard-only gate, not a distinct backend state) | `recording_state == RECORDING` and no case has any accepted raw yet | Every generated case's target state listed in order; "Start Recording" to proceed. Backend has already silently moved to `RECORDING` underneath — harmless, since nothing happens until a capture is pushed. |
| 4 | Default state | same `RECORDING`, before any raw exists | Instructional confirm screen naming the baseline state per the device's states (e.g. "Set POWER=OFF, MODE=COOL, TEMP=24 on the remote, then continue") — no backend call, pure UI gate |
| 5 | Recording | `RECORDING`, mid-flow | Current case's target state name/value; each incoming raw (detected via WS-triggered refetch of `ListCases`) shown as a summary (pulse count, total duration, captured-when) with Accept/Discard buttons (`AcceptRaw`/`DiscardRaw`); prior accepted cases listed below with a "Retry" action (`RetryCase`) for retaking a case the user reconsiders |
| 6 | Analyzing | `ANALYZING` → `FUNCTION_GENERATING` | Progress indicator through the states already broadcast; a generic "refining the generated encoder..." message shown during `FUNCTION_GENERATING` (the repair loop's internal rounds are not surfaced individually — no backend change to expose them) |
| 7 | Testing | `TEST_CASES_GENERATING` → `TESTING` | Per test case (`ListTestCases`): target state, "Transmit" button (`TransmitTestCase`), then "Did it work?" Yes/No (`RecordTestCaseResult`) |
| 8 | Re-record (optional) | back to `RECORDING` after a test failure | Same step-5 UI, re-rendered — not a new screen |
| 9 | Re-analyze (optional) | back to `ANALYZING`/`FUNCTION_GENERATING` | Same step-6 UI, re-rendered |
| 10 | Finishing | `COMPLETED` or `FAILED` | `COMPLETED`: success summary + link into the static detail view. `FAILED`: the session's failure state shown plainly (no separate error-detail endpoint needed — the recording_state itself is the signal) |

Steps 8–9 aren't separate code paths — the step indicator just
highlights differently because `recording_state` cycled back, and the
exact same step-5/6 components render again.

## 4. WebSocket integration

**New frontend-only server action**, `_lib/broadcast-token.ts`:

```ts
"use server";
export async function getBroadcastToken(): Promise<string>
```

Reads the httpOnly access-token cookie server-side (via the existing
session utilities) and returns its raw value — a scoped, deliberate
exception to "client JS never sees the access token," justified by the
WS handler's own documented constraint (browsers can't set an
`Authorization` header on a WebSocket handshake, so the existing backend
endpoint expects the token as a query parameter). The token is already
short-lived (2 minutes); this action can be called again on each
reconnect to fetch a fresh one.

**New hook**, `src/hooks/use-infrared-record-session-broadcast.ts`:

```ts
function useInfraredRecordSessionBroadcast(
  sessionId: string,
  onEvent: () => void,
): { connected: boolean }
```

- Calls `getBroadcastToken()`, opens
  `wss://.../v1/infrared/record-sessions/{sessionId}/broadcast?token=...`.
- Any inbound message calls `onEvent()` (the wizard passes
  `router.refresh` — the broadcast payload itself, per the existing
  backend contract, is a lightweight `InfraredRecordSessionEvent`
  status ping, not a data payload; the actual current cases/coder/test
  cases are re-fetched through the existing REST endpoints on refresh,
  same pattern as `RefreshBoundary` elsewhere in this app, just
  push-triggered instead of timer-triggered).
- On close/error: retries with backoff, re-minting the token via
  `getBroadcastToken()` on each attempt (covering both network drops and
  token expiry).
- If reconnection keeps failing past a bounded number of attempts, falls
  back to `useSmartRefresh`'s interval polling (same 30s-class cadence
  used elsewhere) so the wizard degrades to "slower" rather than "stale
  forever."

## Out of scope

- Any backend change — every step maps to an existing endpoint.
- Surfacing individual repair-loop rounds/scores in the UI (step 6 stays
  at the existing broadcast granularity).
- IR-capability filtering on the node picker in step 1.
- A waveform/timing visualization for raw captures (step 5 uses a
  simple text summary instead).
