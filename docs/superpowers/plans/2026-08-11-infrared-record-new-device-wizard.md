# Infrared Record New Device Wizard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the "Record New Device" wizard — a pre-session form (device + state definitions) that starts a recording session, followed by a live, WebSocket-driven wizard on the session detail page that walks the user through case building, recording, analysis, testing, and finishing, reusing Spec A's list/detail infrastructure throughout.

**Architecture:** `/apps/infrared/record/new` is a standalone client-side 2-step form (never persisted until submit) that calls the existing `POST /infrared/record-sessions` endpoint and redirects into `/apps/infrared/record/[id]`. That detail page now branches: terminal sessions (`is_completed`) keep Spec A's static view unchanged; every other state renders a new `RecordSessionWizard` client component that derives its current step purely from `session.recording_state` plus whether any raw capture exists yet — never from client-stored progress — and refreshes via a native WebSocket hook (with polling fallback) so the view is always correct after a reload. All case/raw/test-case mutation server actions already exist from Spec A and are reused unmodified.

**Tech Stack:** Next.js 16 App Router (Server Components + Server Actions), native browser `WebSocket` (no library), existing design-system components (`Button`, `Input`, `Select`, `Dialog`, `StatusBadge`, `PageHeader`), Go/Echo backend (one additive response-field change).

## Global Constraints

- No backend endpoint gaps except the one explicitly listed in Task 1 (pulse count / duration on raw captures) — every other wizard step maps to an already-existing endpoint/action.
- Wizard state is never persisted client-side. Every render of `RecordSessionWizard` derives its current step from `session.recording_state` (from a fresh server fetch, always the source of truth) plus `cases` (specifically whether any case has `raw.length > 0`). Reloading the tab must always land on the correct step.
- `COMPLETED`/`FAILED` sessions render Spec A's existing static detail view — never the wizard. This is a deliberate resolution of the original design doc's §1/§3 ambiguity: the wizard component is only ever mounted for the 7 non-terminal `recording_state` values, so it never needs its own "step 10" completion/failure screen — the page-level branch routes away before the wizard mounts, and Spec A's already-built overview/cases/coder/test-case sections cover everything a finishing screen would show.
- The design doc's steps 3.5 ("case plan review") and 4 ("confirm default state") are collapsed into a single informational panel with no button — deliberate simplification. There is no backend signal that distinguishes "reviewed the plan" from "about to press the remote"; both are pre-recording states with `recording_state === "RECORDING"` and zero raw captures. The panel just states both things at once and disappears automatically the moment the first raw capture arrives (a real MQTT event, not a client click). On reload before any raw exists, the user sees this panel again — a harmless, non-destructive repeat, not a resumability bug.
- Every mutation the wizard triggers reuses Spec A's `[id]/_lib/actions.ts` and `[id]/_components/RecordCaseList.tsx` / `RecordCaseRawControls.tsx` / `RecordTestCaseList.tsx` verbatim, unmodified. No changes to those files in this plan.
- The WebSocket URL is constructed relative to `window.location` (`${protocol}//${host}/api/v1/infrared/record-sessions/{id}/broadcast?token=...`) — the backend is the single public origin (it reverse-proxies all non-`/api/*` routes to the Next.js server), so no new public env var is needed and `API_BASE_URL` (server-only) is never exposed to the client.
- Only one WebSocket listener per session is allowed by the backend (`gorilla.go`'s `reserve()` rejects a second concurrent registration). The hook must not hammer reconnects — exponential backoff, capped attempts, then fall back to interval polling via `useSmartRefresh`.
- Follow this codebase's existing `useActionState` + `useActionDialog` + `useFirstInvalidField` + `key={dialog.formKey}` pattern for every new create-dialog, exactly as `DeviceTypeForm.tsx` already does.
- Comments stay minimal — only where the WHY is genuinely non-obvious (per this project's established convention).

---

### Task 1: Backend — pulse count and duration on raw capture responses

**Files:**
- Modify: `backend/internal/presentation/http/response/infrared.go:1-9` (imports), `:132-146` (struct + mapper)
- Test: `backend/internal/presentation/http/response/infrared_test.go` (new)

**Interfaces:**
- Consumes: `domainmodels.InfraredStateDeviceRecordRaw.RawData []byte` — already loaded on every read path (`ReadListRawByCaseId`, `ReadRawById`), contains a JSON-marshaled `[]int32` of alternating mark/space microsecond durations (confirmed at `internal/application/infrared/record_session_management/usecase.go:1090`, `:1299`).
- Produces: `InfraredStateDeviceRecordRawResponse.PulseCount int` (`json:"pulse_count"`) and `.DurationUs int` (`json:"duration_us"`) — consumed by Task 6's frontend `WizardCaseRecorder` component and by the type update in Task 2.

The frontend's `InfraredStateDeviceRecordRawResponse` type already has every other field (`id`, `status`, `discarded_reason`, `created_at`, `deleted_at`, `deleted_by`) matching this response 1:1 — Task 2 only adds the two new fields.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/presentation/http/response/infrared_test.go`:

```go
package presentationhttpresponse

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

func TestInfraredStateDeviceRecordRawComputesPulseCountAndDuration(t *testing.T) {
	durations := []int32{9000, 4500, 560, 560, 560, 1690}
	rawData, err := json.Marshal(durations)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}

	model := domainmodels.InfraredStateDeviceRecordRaw{
		Id:        uuid.New(),
		RawData:   rawData,
		Status:    "CAPTURED",
		CreatedAt: time.Now(),
	}

	got := InfraredStateDeviceRecordRaw(model)

	if got.PulseCount != len(durations) {
		t.Fatalf("PulseCount = %d, want %d", got.PulseCount, len(durations))
	}

	wantDuration := 0
	for _, d := range durations {
		wantDuration += int(d)
	}
	if got.DurationUs != wantDuration {
		t.Fatalf("DurationUs = %d, want %d", got.DurationUs, wantDuration)
	}
}

func TestInfraredStateDeviceRecordRawHandlesMalformedRawDataGracefully(t *testing.T) {
	model := domainmodels.InfraredStateDeviceRecordRaw{
		Id:        uuid.New(),
		RawData:   []byte("not json"),
		Status:    "CAPTURED",
		CreatedAt: time.Now(),
	}

	got := InfraredStateDeviceRecordRaw(model)

	if got.PulseCount != 0 || got.DurationUs != 0 {
		t.Fatalf("got PulseCount=%d DurationUs=%d, want 0/0 for malformed raw_data", got.PulseCount, got.DurationUs)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/presentation/http/response/... -run TestInfraredStateDeviceRecordRaw -v`
Expected: FAIL — `got.PulseCount`/`got.DurationUs` undefined (fields don't exist yet), a compile error.

- [ ] **Step 3: Add the fields and compute them in the mapper**

In `backend/internal/presentation/http/response/infrared.go`, add `"encoding/json"` to the import block (currently just `"time"`, the two domain-model imports, and `"github.com/google/uuid"`):

```go
import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	"github.com/google/uuid"
)
```

Replace lines 132-146 with:

```go
type InfraredStateDeviceRecordRawResponse struct {
	Id              string  `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	Status          string  `json:"status" example:"CAPTURED"`
	DiscardedReason *string `json:"discarded_reason,omitempty" example:"pressed the wrong button"`
	PulseCount      int     `json:"pulse_count" example:"48"`
	DurationUs      int     `json:"duration_us" example:"862000"`
	PipelineAuditResponse
}

// InfraredStateDeviceRecordRaw unmarshals RawData (mark/space microsecond
// durations, JSON-encoded []int32) to summarize a capture without exposing
// the raw timing array to the client. A malformed payload degrades to 0/0
// rather than failing the whole response - raw_data is NOT NULL and always
// written by CaptureIrRaw, so this only matters for hand-corrupted rows.
func InfraredStateDeviceRecordRaw(model domainmodels.InfraredStateDeviceRecordRaw) InfraredStateDeviceRecordRawResponse {
	var durations []int32
	_ = json.Unmarshal(model.RawData, &durations)

	durationUs := 0
	for _, d := range durations {
		durationUs += int(d)
	}

	return InfraredStateDeviceRecordRawResponse{
		Id:                    UUIDString(model.Id),
		Status:                string(model.Status),
		DiscardedReason:       model.DiscardedReason,
		PulseCount:            len(durations),
		DurationUs:            durationUs,
		PipelineAuditResponse: PipelineAudit(model.CreatedAt, model.DeletedAt, model.DeletedBy),
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go test ./internal/presentation/http/response/... -v`
Expected: PASS, including the two new tests and every existing test in the package (`llm_config_test.go`).

- [ ] **Step 5: Run the full backend build and test suite**

Run: `cd backend && go build ./... && go test ./...`
Expected: PASS, no other package references the old field set in a way that breaks.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/presentation/http/response/infrared.go backend/internal/presentation/http/response/infrared_test.go
git commit -m "feat(infrared): expose pulse count and duration on raw capture responses"
```

---

### Task 2: Frontend — API type update, shared Stepper and StringListEditor components

**Files:**
- Modify: `frontend/src/lib/api/infrared.ts` (one interface)
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/_components/WizardStepper.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/_components/StringListEditor.tsx`
- Test: manual (component library task, exercised end-to-end by Tasks 4 and 6)

**Interfaces:**
- Produces: `InfraredStateDeviceRecordRawResponse` gains `pulse_count: number` and `duration_us: number`, matching Task 1's backend response exactly. `WizardStepper({steps, currentIndex}: {steps: readonly string[]; currentIndex: number})` — default export. `StringListEditor({values, onChange, placeholder}: {values: readonly string[]; onChange: (values: string[]) => void; placeholder?: string})` — default export, controlled component.
- Consumes: `lucide-react` (`Check`, `Plus`, `X` icons — already a dependency, used throughout this codebase), `Button`/`Input` from `@/components/ui/button` / `@/components/ui/input` (read in full during planning, prop surfaces confirmed stable).

- [ ] **Step 1: Update the raw response type**

In `frontend/src/lib/api/infrared.ts`, change:

```ts
export interface InfraredStateDeviceRecordRawResponse {
  id: string;
  status: "CAPTURED" | "ACCEPTED" | "DISCARDED";
  discarded_reason?: string;
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}
```

to:

```ts
export interface InfraredStateDeviceRecordRawResponse {
  id: string;
  status: "CAPTURED" | "ACCEPTED" | "DISCARDED";
  discarded_reason?: string;
  pulse_count: number;
  duration_us: number;
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}
```

- [ ] **Step 2: Verify the type change compiles**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS (this is a purely additive field; nothing destructures `InfraredStateDeviceRecordRawResponse` exhaustively today).

- [ ] **Step 3: Create the stepper component**

Create `frontend/src/app/(authenticated)/apps/infrared/record/_components/WizardStepper.tsx`:

```tsx
import { Check } from "lucide-react";

export default function WizardStepper({
  steps,
  currentIndex,
}: {
  steps: readonly string[];
  currentIndex: number;
}) {
  return (
    <ol className="flex flex-wrap gap-x-6 gap-y-3 text-sm">
      {steps.map((label, index) => {
        const done = index < currentIndex;
        const active = index === currentIndex;
        return (
          <li key={label} className="flex items-center gap-2">
            <span
              className={`flex size-6 shrink-0 items-center justify-center rounded-full border text-xs font-semibold ${
                done
                  ? "border-success bg-success text-surface"
                  : active
                    ? "border-primary text-primary"
                    : "border-border text-muted-foreground"
              }`}
            >
              {done ? (
                <Check aria-hidden="true" className="size-3.5" />
              ) : (
                index + 1
              )}
            </span>
            <span
              className={
                active
                  ? "text-foreground font-semibold"
                  : "text-foreground/70"
              }
            >
              {label}
            </span>
          </li>
        );
      })}
    </ol>
  );
}
```

- [ ] **Step 4: Create the string-list editor**

Create `frontend/src/app/(authenticated)/apps/infrared/record/_components/StringListEditor.tsx`:

```tsx
"use client";

import { Plus, X } from "lucide-react";

import Button from "@/components/ui/button";
import Input from "@/components/ui/input";

export default function StringListEditor({
  values,
  onChange,
  placeholder,
}: {
  values: readonly string[];
  onChange: (values: string[]) => void;
  placeholder?: string;
}) {
  return (
    <div className="space-y-2">
      {values.map((value, index) => (
        <div key={index} className="flex items-center gap-2">
          <Input
            value={value}
            placeholder={placeholder}
            onChange={(event) => {
              const next = [...values];
              next[index] = event.target.value;
              onChange(next);
            }}
          />
          <Button
            type="button"
            variant="secondary"
            aria-label="Remove option"
            onClick={() => onChange(values.filter((_, i) => i !== index))}
          >
            <X aria-hidden="true" className="size-4" />
          </Button>
        </div>
      ))}
      <Button
        type="button"
        variant="secondary"
        onClick={() => onChange([...values, ""])}
      >
        <Plus aria-hidden="true" className="mr-1.5 size-4" />
        Add option
      </Button>
    </div>
  );
}
```

- [ ] **Step 5: Lint, format, typecheck, build**

Run: `cd frontend && npm run lint && npx prettier --check . && npx tsc --noEmit && npm run build`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/api/infrared.ts frontend/src/app/'(authenticated)'/apps/infrared/record/_components/WizardStepper.tsx frontend/src/app/'(authenticated)'/apps/infrared/record/_components/StringListEditor.tsx
git commit -m "feat(infrared): add raw capture metadata field and shared wizard UI primitives"
```

---

### Task 3: Frontend — pre-session server actions (`/record/new/_lib`)

**Files:**
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/new/_lib/state.ts`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/new/_lib/actions.ts`

**Interfaces:**
- Consumes: `startInfraredRecordSession`, `createInfraredDeviceType`, `createInfraredState`, `listInfraredStates` from `@/lib/api/infrared` (all pre-existing, signatures confirmed in Task-planning read of `lib/api/infrared.ts`); `requireSessionContext` from `@/lib/session`; `ApiError` from `@/lib/api/client`.
- Produces: `NewRecordSessionActionState` (= `ActionState<string>`), `WizardCreatedItem {id: string; name: string}`, `WizardCreateActionState` (= `ActionState<string> & {created?: WizardCreatedItem}`), `EMPTY_NEW_RECORD_SESSION_STATE`, `EMPTY_WIZARD_CREATE_STATE` — consumed by Task 4's form and dialogs. `startRecordSessionAction(prevState, formData)` reads a single hidden `payload` field containing a JSON-serialized `StartRecordSessionRequest`; on success it redirects to `/apps/infrared/record/{id}`, it never returns a "success" state. `createDeviceTypeForWizardAction`/`createStateForWizardAction(prevState, formData)` return `WizardCreateActionState` with `created` populated on success. `listStatesForDeviceTypeAction(deviceTypeId: string): Promise<InfraredStateResponse[]>` is a plain async server action (not form-bound) callable directly from client code.

- [ ] **Step 1: Create the state module**

Create `frontend/src/app/(authenticated)/apps/infrared/record/new/_lib/state.ts`:

```ts
import type { ActionState } from "@/lib/forms/action-state";

export type NewRecordSessionActionState = ActionState<string>;

export const EMPTY_NEW_RECORD_SESSION_STATE: NewRecordSessionActionState = {
  status: "idle",
};

export interface WizardCreatedItem {
  id: string;
  name: string;
}

export type WizardCreateActionState = ActionState<string> & {
  created?: WizardCreatedItem;
};

export const EMPTY_WIZARD_CREATE_STATE: WizardCreateActionState = {
  status: "idle",
};
```

- [ ] **Step 2: Create the actions module**

Create `frontend/src/app/(authenticated)/apps/infrared/record/new/_lib/actions.ts`:

```ts
"use server";

import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/client";
import {
  createInfraredDeviceType,
  createInfraredState,
  listInfraredStates,
  startInfraredRecordSession,
  type InfraredStateResponse,
  type InfraredStateType,
  type StartRecordSessionRequest,
} from "@/lib/api/infrared";
import { requireSessionContext } from "@/lib/session";

import type {
  NewRecordSessionActionState,
  WizardCreateActionState,
} from "./state";

function permissionDenied(): NewRecordSessionActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): NewRecordSessionActionState {
  if (error instanceof ApiError) {
    return { status: "error", title: error.title, message: error.message };
  }
  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

export async function listStatesForDeviceTypeAction(
  deviceTypeId: string,
): Promise<InfraredStateResponse[]> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_reference:get") || !deviceTypeId) {
    return [];
  }
  return listInfraredStates(deviceTypeId);
}

export async function startRecordSessionAction(
  _previousState: NewRecordSessionActionState,
  formData: FormData,
): Promise<NewRecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:add")) {
    return permissionDenied();
  }

  let payload: StartRecordSessionRequest;
  try {
    payload = JSON.parse(
      String(formData.get("payload") ?? ""),
    ) as StartRecordSessionRequest;
  } catch {
    return {
      status: "error",
      title: "Something went wrong",
      message: "Could not read the form. Please try again.",
    };
  }

  if (
    !payload.node_id ||
    !payload.infrared_device_type_id ||
    !payload.brand?.trim() ||
    !payload.model?.trim() ||
    !payload.definitions?.length
  ) {
    return {
      status: "error",
      title: "Check the form",
      message: "Fill in every required field before starting.",
    };
  }

  let created;
  try {
    created = await startInfraredRecordSession(payload);
  } catch (error) {
    return actionError(error);
  }

  redirect(`/apps/infrared/record/${created.id}`);
}

export async function createDeviceTypeForWizardAction(
  _previousState: WizardCreateActionState,
  formData: FormData,
): Promise<WizardCreateActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_reference:add")) {
    return permissionDenied();
  }

  const name = String(formData.get("name") ?? "").trim();
  if (!name) {
    return {
      status: "error",
      title: "Check the device type",
      message: "Complete the required field.",
      fieldErrors: { name: "This field is required." },
    };
  }

  let created;
  try {
    created = await createInfraredDeviceType(name);
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Device type created",
    message: `${name} is ready to use.`,
    created: { id: created.id, name },
  };
}

export async function createStateForWizardAction(
  _previousState: WizardCreateActionState,
  formData: FormData,
): Promise<WizardCreateActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_reference:add")) {
    return permissionDenied();
  }

  const deviceTypeId = String(formData.get("device_type_id") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const type = String(formData.get("type") ?? "");

  const fieldErrors: Record<string, string> = {};
  if (!name) fieldErrors.name = "This field is required.";
  if (type !== "ENUM" && type !== "RANGE") {
    fieldErrors.type = "Choose a type.";
  }
  if (Object.keys(fieldErrors).length) {
    return {
      status: "error",
      title: "Check the state",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }

  let created;
  try {
    created = await createInfraredState(deviceTypeId, {
      name,
      type: type as InfraredStateType,
    });
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "State created",
    message: `${name} is ready to use.`,
    created: { id: created.id, name },
  };
}
```

- [ ] **Step 3: Typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS. (No test framework exists for server actions in this codebase — Spec A's equivalent actions in `[id]/_lib/actions.ts` have no unit tests either; this file follows the same precedent and is exercised end-to-end via Task 4's form.)

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/'(authenticated)'/apps/infrared/record/new/_lib/state.ts frontend/src/app/'(authenticated)'/apps/infrared/record/new/_lib/actions.ts
git commit -m "feat(infrared): add server actions for starting a record session and inline reference creation"
```

---

### Task 4: Frontend — pre-session form page (`/record/new`, wizard steps 1-2)

**Files:**
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/new/page.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/new/_components/DeviceTypeCreateDialog.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/new/_components/StateCreateDialog.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/new/_components/NewRecordSessionForm.tsx`

**Interfaces:**
- Consumes: Task 2's `WizardStepper`/`StringListEditor`, Task 3's actions/state, `listAllNodes` from `@/lib/api/nodes`, `listInfraredDeviceTypes` from `@/lib/api/infrared`, `requirePermission` from `@/lib/session`, `useActionDialog`/`useFirstInvalidField` (both pre-existing, signatures confirmed).
- Produces: nothing consumed by later tasks — this route is a leaf. Task 7 links to it.

- [ ] **Step 1: Create the page**

Create `frontend/src/app/(authenticated)/apps/infrared/record/new/page.tsx`:

```tsx
import type { Metadata } from "next";

import PageHeader from "@/components/ui/page-header";
import { listInfraredDeviceTypes } from "@/lib/api/infrared";
import { listAllNodes } from "@/lib/api/nodes";
import { requirePermission } from "@/lib/session";

import NewRecordSessionForm from "./_components/NewRecordSessionForm";

export const metadata: Metadata = { title: "Record New Device — Mate Things" };

export default async function NewRecordSessionPage() {
  const { permissions } = await requirePermission(
    "infrared_record_session:add",
  );

  const [nodes, deviceTypes] = await Promise.all([
    permissions.has("node:get") ? listAllNodes() : Promise.resolve([]),
    permissions.has("infrared_reference:get")
      ? listInfraredDeviceTypes()
      : Promise.resolve([]),
  ]);

  return (
    <main className="mx-auto w-full max-w-3xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Record New Device"
        description="Capture a remote's infrared signals to teach the Infrared app how to control this device."
      />
      <NewRecordSessionForm nodes={nodes} initialDeviceTypes={deviceTypes} />
    </main>
  );
}
```

- [ ] **Step 2: Create the device-type inline-creation dialog**

Create `frontend/src/app/(authenticated)/apps/infrared/record/new/_components/DeviceTypeCreateDialog.tsx`:

```tsx
"use client";

import { useActionState, useEffect, useId, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useActionDialog } from "@/hooks/use-action-dialog";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";

import { createDeviceTypeForWizardAction } from "../_lib/actions";
import {
  EMPTY_WIZARD_CREATE_STATE,
  type WizardCreatedItem,
} from "../_lib/state";

export default function DeviceTypeCreateDialog({
  onCreated,
}: {
  onCreated: (item: WizardCreatedItem) => void;
}) {
  const [state, setState] = useState(EMPTY_WIZARD_CREATE_STATE);
  const dialog = useActionDialog({
    state,
    onSuccess: () => {
      if (state.created) {
        onCreated(state.created);
      }
    },
  });

  return (
    <>
      <Button
        type="button"
        variant="secondary"
        onClick={() => dialog.setOpen(true)}
      >
        New device type
      </Button>
      <DeviceTypeCreateContent
        key={dialog.formKey}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
      />
    </>
  );
}

function DeviceTypeCreateContent({
  open,
  onClose,
  onStateChange,
}: {
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_WIZARD_CREATE_STATE) => void;
}) {
  const [state, action, pending] = useActionState(
    createDeviceTypeForWizardAction,
    EMPTY_WIZARD_CREATE_STATE,
  );
  const fieldId = useId();
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  useEffect(() => onStateChange(state), [onStateChange, state]);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Create device type"
      variant="sheet"
      dismissible={!pending}
    >
      <form
        ref={formRef}
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <div>
          <Label htmlFor={`${fieldId}-name`}>Name</Label>
          <Input
            id={`${fieldId}-name`}
            name="name"
            placeholder="Air Conditioner"
            required
            aria-invalid={Boolean(state.fieldErrors?.name)}
            aria-describedby={
              state.fieldErrors?.name ? `${fieldId}-name-error` : undefined
            }
          />
          <FieldError id={`${fieldId}-name-error`}>
            {state.fieldErrors?.name}
          </FieldError>
        </div>
        <ActionMessage state={state} />
        <div className="flex flex-wrap justify-end gap-2">
          <Button
            type="button"
            variant="secondary"
            disabled={pending}
            onClick={onClose}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Creating…" : "Create device type"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
```

- [ ] **Step 3: Create the state inline-creation dialog**

Create `frontend/src/app/(authenticated)/apps/infrared/record/new/_components/StateCreateDialog.tsx`:

```tsx
"use client";

import { useActionState, useEffect, useId, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import Select from "@/components/ui/select";
import { useActionDialog } from "@/hooks/use-action-dialog";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import type { InfraredStateType } from "@/lib/api/infrared";

import { createStateForWizardAction } from "../_lib/actions";
import {
  EMPTY_WIZARD_CREATE_STATE,
  type WizardCreatedItem,
} from "../_lib/state";

export default function StateCreateDialog({
  deviceTypeId,
  onCreated,
}: {
  deviceTypeId: string;
  onCreated: (item: WizardCreatedItem, type: InfraredStateType) => void;
}) {
  const [state, setState] = useState(EMPTY_WIZARD_CREATE_STATE);
  const [type, setType] = useState<InfraredStateType>("ENUM");
  const dialog = useActionDialog({
    state,
    onSuccess: () => {
      if (state.created) {
        onCreated(state.created, type);
      }
    },
  });

  return (
    <>
      <Button
        type="button"
        variant="secondary"
        disabled={!deviceTypeId}
        onClick={() => dialog.setOpen(true)}
      >
        New state
      </Button>
      <StateCreateContent
        key={dialog.formKey}
        deviceTypeId={deviceTypeId}
        open={dialog.open}
        onClose={dialog.reset}
        onStateChange={setState}
        onTypeChange={setType}
      />
    </>
  );
}

function StateCreateContent({
  deviceTypeId,
  open,
  onClose,
  onStateChange,
  onTypeChange,
}: {
  deviceTypeId: string;
  open: boolean;
  onClose: () => void;
  onStateChange: (state: typeof EMPTY_WIZARD_CREATE_STATE) => void;
  onTypeChange: (type: InfraredStateType) => void;
}) {
  const [state, action, pending] = useActionState(
    createStateForWizardAction,
    EMPTY_WIZARD_CREATE_STATE,
  );
  const fieldId = useId();
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  useEffect(() => onStateChange(state), [onStateChange, state]);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Create state"
      variant="sheet"
      dismissible={!pending}
    >
      <form
        ref={formRef}
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="device_type_id" value={deviceTypeId} />
        <div>
          <Label htmlFor={`${fieldId}-name`}>Name</Label>
          <Input
            id={`${fieldId}-name`}
            name="name"
            placeholder="Mode"
            required
            aria-invalid={Boolean(state.fieldErrors?.name)}
            aria-describedby={
              state.fieldErrors?.name ? `${fieldId}-name-error` : undefined
            }
          />
          <FieldError id={`${fieldId}-name-error`}>
            {state.fieldErrors?.name}
          </FieldError>
        </div>
        <div>
          <Label htmlFor={`${fieldId}-type`}>Type</Label>
          <Select
            id={`${fieldId}-type`}
            name="type"
            defaultValue="ENUM"
            onChange={(event) =>
              onTypeChange(event.target.value as InfraredStateType)
            }
          >
            <option value="ENUM">Enum (fixed options)</option>
            <option value="RANGE">Range (min/max/step)</option>
          </Select>
          <FieldError id={`${fieldId}-type-error`}>
            {state.fieldErrors?.type}
          </FieldError>
        </div>
        <ActionMessage state={state} />
        <div className="flex flex-wrap justify-end gap-2">
          <Button
            type="button"
            variant="secondary"
            disabled={pending}
            onClick={onClose}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Creating…" : "Create state"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
```

- [ ] **Step 4: Create the two-step form**

Create `frontend/src/app/(authenticated)/apps/infrared/record/new/_components/NewRecordSessionForm.tsx`:

```tsx
"use client";

import { useActionState, useEffect, useId, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import Select from "@/components/ui/select";
import type {
  InfraredDeviceTypeResponse,
  InfraredStateResponse,
  InfraredStateType,
  StartRecordSessionDefinition,
} from "@/lib/api/infrared";
import type { NodeResponse } from "@/lib/api/nodes";

import {
  listStatesForDeviceTypeAction,
  startRecordSessionAction,
} from "../_lib/actions";
import { EMPTY_NEW_RECORD_SESSION_STATE } from "../_lib/state";
import StringListEditor from "../../_components/StringListEditor";
import WizardStepper from "../../_components/WizardStepper";
import DeviceTypeCreateDialog from "./DeviceTypeCreateDialog";
import StateCreateDialog from "./StateCreateDialog";

const PRE_SESSION_STEPS = ["Device details", "State values"] as const;

interface StateDraft {
  stateId: string;
  name: string;
  type: InfraredStateType;
  options: string[];
  minimum: string;
  maximum: string;
  step: string;
}

function draftFromState(state: InfraredStateResponse): StateDraft {
  return {
    stateId: state.id,
    name: state.name,
    type: state.type,
    options: state.type === "ENUM" ? [""] : [],
    minimum: "",
    maximum: "",
    step: "",
  };
}

export default function NewRecordSessionForm({
  nodes,
  initialDeviceTypes,
}: {
  nodes: readonly NodeResponse[];
  initialDeviceTypes: readonly InfraredDeviceTypeResponse[];
}) {
  const [wizardStep, setWizardStep] = useState<1 | 2>(1);
  const [deviceTypes, setDeviceTypes] = useState(initialDeviceTypes);
  const [nodeId, setNodeId] = useState("");
  const [deviceTypeId, setDeviceTypeId] = useState("");
  const [brand, setBrand] = useState("");
  const [model, setModel] = useState("");
  const [drafts, setDrafts] = useState<StateDraft[]>([]);
  const [statesLoading, setStatesLoading] = useState(false);
  const [stepError, setStepError] = useState<string | null>(null);
  const fieldId = useId();

  const [state, formAction, pending] = useActionState(
    startRecordSessionAction,
    EMPTY_NEW_RECORD_SESSION_STATE,
  );

  useEffect(() => {
    if (wizardStep !== 2 || !deviceTypeId) {
      return;
    }
    let cancelled = false;
    setStatesLoading(true);
    listStatesForDeviceTypeAction(deviceTypeId)
      .then((states) => {
        if (!cancelled) {
          setDrafts(states.map(draftFromState));
        }
      })
      .finally(() => {
        if (!cancelled) {
          setStatesLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [wizardStep, deviceTypeId]);

  function goToStep2() {
    if (!nodeId || !deviceTypeId || !brand.trim() || !model.trim()) {
      setStepError("Fill in every field before continuing.");
      return;
    }
    setStepError(null);
    setWizardStep(2);
  }

  function buildDefinitions(): StartRecordSessionDefinition[] {
    return drafts.map((draft) =>
      draft.type === "ENUM"
        ? {
            infrared_state_id: draft.stateId,
            options: draft.options.map((o) => o.trim()).filter(Boolean),
          }
        : {
            infrared_state_id: draft.stateId,
            minimum: Number(draft.minimum),
            maximum: Number(draft.maximum),
            step: Number(draft.step),
          },
    );
  }

  function handleSubmit(formData: FormData) {
    if (drafts.length === 0) {
      setStepError("Add at least one state value before starting.");
      return;
    }
    const invalidEnum = drafts.some(
      (d) =>
        d.type === "ENUM" && d.options.filter((o) => o.trim()).length === 0,
    );
    const invalidRange = drafts.some(
      (d) =>
        d.type === "RANGE" &&
        (d.minimum === "" || d.maximum === "" || d.step === ""),
    );
    if (invalidEnum || invalidRange) {
      setStepError("Complete every state's values before starting.");
      return;
    }
    setStepError(null);
    formData.set(
      "payload",
      JSON.stringify({
        node_id: nodeId,
        infrared_device_type_id: deviceTypeId,
        brand: brand.trim(),
        model: model.trim(),
        definitions: buildDefinitions(),
      }),
    );
    formAction(formData);
  }

  return (
    <div className="space-y-6">
      <WizardStepper steps={PRE_SESSION_STEPS} currentIndex={wizardStep - 1} />
      {wizardStep === 1 ? (
        <div className="space-y-4">
          <div>
            <Label htmlFor={`${fieldId}-node`}>Node</Label>
            <Select
              id={`${fieldId}-node`}
              value={nodeId}
              onChange={(event) => setNodeId(event.target.value)}
            >
              <option value="">Select a node</option>
              {nodes.map((node) => (
                <option key={node.id} value={node.id}>
                  {node.name || node.device_id}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <Label htmlFor={`${fieldId}-device-type`}>Device type</Label>
            <div className="flex items-center gap-2">
              <Select
                id={`${fieldId}-device-type`}
                value={deviceTypeId}
                onChange={(event) => setDeviceTypeId(event.target.value)}
              >
                <option value="">Select a device type</option>
                {deviceTypes.map((dt) => (
                  <option key={dt.id} value={dt.id}>
                    {dt.name}
                  </option>
                ))}
              </Select>
              <DeviceTypeCreateDialog
                onCreated={(item) => {
                  setDeviceTypes((prev) => [
                    ...prev,
                    {
                      id: item.id,
                      name: item.name,
                      created_at: new Date().toISOString(),
                    },
                  ]);
                  setDeviceTypeId(item.id);
                }}
              />
            </div>
          </div>
          <div>
            <Label htmlFor={`${fieldId}-brand`}>Brand</Label>
            <Input
              id={`${fieldId}-brand`}
              value={brand}
              onChange={(event) => setBrand(event.target.value)}
            />
          </div>
          <div>
            <Label htmlFor={`${fieldId}-model`}>Model</Label>
            <Input
              id={`${fieldId}-model`}
              value={model}
              onChange={(event) => setModel(event.target.value)}
            />
          </div>
          {stepError ? (
            <p className="text-critical text-sm">{stepError}</p>
          ) : null}
          <div className="flex justify-end">
            <Button type="button" onClick={goToStep2}>
              Continue
            </Button>
          </div>
        </div>
      ) : (
        <form action={handleSubmit} className="space-y-4">
          {statesLoading ? (
            <p className="text-foreground/70 text-sm">Loading states…</p>
          ) : (
            drafts.map((draft, index) => (
              <div
                key={draft.stateId}
                className="border-border rounded-2xl border p-4"
              >
                <p className="font-semibold">{draft.name}</p>
                {draft.type === "ENUM" ? (
                  <div className="mt-2">
                    <StringListEditor
                      values={draft.options}
                      placeholder="Option value"
                      onChange={(options) => {
                        const next = [...drafts];
                        next[index] = { ...draft, options };
                        setDrafts(next);
                      }}
                    />
                  </div>
                ) : (
                  <div className="mt-2 grid grid-cols-3 gap-2">
                    <Input
                      type="number"
                      placeholder="Minimum"
                      value={draft.minimum}
                      onChange={(event) => {
                        const next = [...drafts];
                        next[index] = { ...draft, minimum: event.target.value };
                        setDrafts(next);
                      }}
                    />
                    <Input
                      type="number"
                      placeholder="Maximum"
                      value={draft.maximum}
                      onChange={(event) => {
                        const next = [...drafts];
                        next[index] = { ...draft, maximum: event.target.value };
                        setDrafts(next);
                      }}
                    />
                    <Input
                      type="number"
                      placeholder="Step"
                      value={draft.step}
                      onChange={(event) => {
                        const next = [...drafts];
                        next[index] = { ...draft, step: event.target.value };
                        setDrafts(next);
                      }}
                    />
                  </div>
                )}
              </div>
            ))
          )}
          <StateCreateDialog
            deviceTypeId={deviceTypeId}
            onCreated={(item, type) => {
              setDrafts((prev) => [
                ...prev,
                {
                  stateId: item.id,
                  name: item.name,
                  type,
                  options: type === "ENUM" ? [""] : [],
                  minimum: "",
                  maximum: "",
                  step: "",
                },
              ]);
            }}
          />
          <ActionMessage state={state} />
          {stepError ? (
            <p className="text-critical text-sm">{stepError}</p>
          ) : null}
          <div className="flex justify-between">
            <Button
              type="button"
              variant="secondary"
              onClick={() => setWizardStep(1)}
            >
              Back
            </Button>
            <Button type="submit" disabled={pending}>
              {pending ? "Starting…" : "Start recording"}
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}
```

- [ ] **Step 5: Lint, format, typecheck, build**

Run: `cd frontend && npm run lint && npx prettier --check . && npx tsc --noEmit && npm run build`
Expected: PASS.

- [ ] **Step 6: Manual browser verification**

Start the dev stack (backend + frontend, per this repo's existing run instructions — use `localhost`, not `127.0.0.1`, per this project's known hydration quirk). Log in with an account holding `infrared_record_session:add`, `infrared_reference:add`, `infrared_reference:get`, `node:get`. Navigate to `/apps/infrared/record/new`:
- Confirm the node and device-type selects populate.
- Create a new device type inline via "New device type" — confirm it appears in the select immediately (no page reload) and gets auto-selected.
- Fill brand/model, click Continue — confirm step 2 loads that device type's states.
- Create a new state inline via "New state" (both ENUM and RANGE) — confirm each appears in the step-2 list immediately.
- Fill in ENUM options and RANGE min/max/step, click "Start recording" — confirm redirect to `/apps/infrared/record/{id}` and a new session was created (check the list page).

- [ ] **Step 7: Commit**

```bash
git add frontend/src/app/'(authenticated)'/apps/infrared/record/new/
git commit -m "feat(infrared): add pre-session device and state definition form"
```

---

### Task 5: Frontend — broadcast token action and WebSocket hook

**Files:**
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_lib/broadcast-token.ts`
- Create: `frontend/src/hooks/use-infrared-record-session-broadcast.ts`

**Interfaces:**
- Consumes: `cookies` from `next/headers`, `ACCESS_TOKEN_COOKIE` from `@/lib/session/cookies`, `requireSessionContext` from `@/lib/session` — same pattern `apiFetch`'s Authorization-header builder already uses.
- Produces: `getBroadcastToken(): Promise<string | null>` — a re-callable server action (fresh token read on every call, since the hook calls it on every (re)connect attempt). `useInfraredRecordSessionBroadcast(sessionId: string, onEvent: () => void, getToken: () => Promise<string | null>): {connected: boolean; exhausted: boolean}` — consumed by Task 6's `RecordSessionWizard`.

- [ ] **Step 1: Create the broadcast-token server action**

Create `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_lib/broadcast-token.ts`:

```ts
"use server";

import { cookies } from "next/headers";

import { requireSessionContext } from "@/lib/session";
import { ACCESS_TOKEN_COOKIE } from "@/lib/session/cookies";

// The WebSocket handshake can't set an Authorization header, so the token
// has to travel as a query param the client can read - this is the one
// deliberate, narrow exception to "client JS never sees the access token".
// It's re-callable on every (re)connect attempt so a refreshed cookie is
// always picked up.
export async function getBroadcastToken(): Promise<string | null> {
  await requireSessionContext();
  const cookieStore = await cookies();
  return cookieStore.get(ACCESS_TOKEN_COOKIE)?.value ?? null;
}
```

- [ ] **Step 2: Create the WebSocket hook**

Create `frontend/src/hooks/use-infrared-record-session-broadcast.ts`:

```ts
"use client";

import { useEffect, useRef, useState } from "react";

const MAX_RECONNECT_ATTEMPTS = 5;
const BASE_RECONNECT_DELAY_MS = 1000;

// Only one WebSocket listener is allowed per session on the backend
// (gorilla.go's reserve() rejects a second concurrent registration) - the
// exponential backoff here exists so a stuck/slow-to-release server-side
// listener isn't hammered with reconnect attempts. After MAX_RECONNECT_
// ATTEMPTS the caller is expected to fall back to interval polling.
export function useInfraredRecordSessionBroadcast(
  sessionId: string,
  onEvent: () => void,
  getToken: () => Promise<string | null>,
): { connected: boolean; exhausted: boolean } {
  const [connected, setConnected] = useState(false);
  const [exhausted, setExhausted] = useState(false);
  const onEventRef = useRef(onEvent);
  const getTokenRef = useRef(getToken);
  const attemptRef = useRef(0);

  useEffect(() => {
    onEventRef.current = onEvent;
    getTokenRef.current = getToken;
  }, [onEvent, getToken]);

  useEffect(() => {
    let cancelled = false;
    let socket: WebSocket | null = null;
    let retryTimeout: number | undefined;

    async function connect() {
      const token = await getTokenRef.current();
      if (cancelled || !token) {
        return;
      }

      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const url = `${protocol}//${window.location.host}/api/v1/infrared/record-sessions/${sessionId}/broadcast?token=${encodeURIComponent(token)}`;
      socket = new WebSocket(url);

      socket.onopen = () => {
        if (cancelled) return;
        attemptRef.current = 0;
        setConnected(true);
        setExhausted(false);
      };

      socket.onmessage = () => {
        if (!cancelled) {
          onEventRef.current();
        }
      };

      socket.onclose = () => {
        if (cancelled) return;
        setConnected(false);
        if (attemptRef.current >= MAX_RECONNECT_ATTEMPTS) {
          setExhausted(true);
          return;
        }
        const delay = BASE_RECONNECT_DELAY_MS * 2 ** attemptRef.current;
        attemptRef.current += 1;
        retryTimeout = window.setTimeout(connect, delay);
      };

      socket.onerror = () => {
        socket?.close();
      };
    }

    connect();

    return () => {
      cancelled = true;
      if (retryTimeout) {
        window.clearTimeout(retryTimeout);
      }
      socket?.close();
    };
  }, [sessionId]);

  return { connected, exhausted };
}
```

- [ ] **Step 3: Typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS. (No unit test framework exists for hooks in this codebase — this hook is exercised end-to-end in Task 6's manual browser verification, where the WS connection, a forced disconnect/reconnect, and the polling fallback are all observed directly.)

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/'(authenticated)'/apps/infrared/record/'[id]'/_lib/broadcast-token.ts frontend/src/hooks/use-infrared-record-session-broadcast.ts
git commit -m "feat(infrared): add broadcast token action and record session WebSocket hook"
```

---

### Task 6: Frontend — live wizard component and detail-page routing branch

**Files:**
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/WizardCaseRecorder.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/RecordSessionWizard.tsx`
- Modify: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/page.tsx`

**Interfaces:**
- Consumes: Task 2's `WizardStepper`, Task 5's `getBroadcastToken`/`useInfraredRecordSessionBroadcast`, `useSmartRefresh` from `@/hooks/use-smart-refresh` (pre-existing, signature confirmed), Spec A's `RecordCaseList`, `RecordCaseRawControls`, `RecordTestCaseList` (all reused unmodified), `InfraredRecordSessionResponse`/`InfraredStateDeviceRecordCaseResponse`/`InfraredTestCaseResponse` types from `@/lib/api/infrared`.
- Produces: `RecordSessionWizard({session, cases, testCases, canMutate}) ` — default export, mounted by `page.tsx` for every non-terminal session.

- [ ] **Step 1: Create the active-case recording panel**

Create `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/WizardCaseRecorder.tsx`:

```tsx
import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge from "@/components/ui/status-badge";
import type { InfraredStateDeviceRecordCaseResponse } from "@/lib/api/infrared";

import RecordCaseList from "./RecordCaseList";
import RecordCaseRawControls from "./RecordCaseRawControls";

function formatDuration(durationUs: number): string {
  return `${Math.round(durationUs / 1000)}ms`;
}

export default function WizardCaseRecorder({
  cases,
  canMutate,
}: {
  cases: readonly InfraredStateDeviceRecordCaseResponse[];
  canMutate: boolean;
}) {
  const activeCase = cases.find((c) => c.status === "ACTIVE");
  // The active case is rendered prominently above with its own controls -
  // excluding it from the roster below avoids showing its Accept/Discard
  // buttons twice for the same raw.
  const rosterCases = cases.filter((c) => c.status !== "ACTIVE");

  return (
    <div className="space-y-4">
      {activeCase ? (
        <div className="border-border space-y-3 rounded-2xl border p-6">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div>
              <p className="text-foreground/70 text-xs font-semibold tracking-wide uppercase">
                Press the remote now
              </p>
              <p className="mt-1 font-semibold">
                {activeCase.description || `Step ${activeCase.step}`}
              </p>
              <p className="text-foreground/70 text-sm">
                {activeCase.states.map((s) => s.state_value).join(", ")}
              </p>
            </div>
            <StatusBadge variant="info">Waiting for capture</StatusBadge>
          </div>
          {activeCase.raw.length ? (
            <ul className="space-y-2">
              {activeCase.raw.map((raw) => (
                <li
                  key={raw.id}
                  className="border-border flex flex-wrap items-center justify-between gap-2 rounded-xl border p-3 text-sm"
                >
                  <span>
                    {raw.pulse_count} pulses ·{" "}
                    {formatDuration(raw.duration_us)} ·{" "}
                    <LocalDateTime value={raw.created_at} />
                  </span>
                  <RecordCaseRawControls
                    caseId={activeCase.id}
                    raw={raw}
                    canMutate={canMutate}
                  />
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-foreground/70 text-sm">
              No captures yet — press the button on the remote that matches
              this state.
            </p>
          )}
        </div>
      ) : null}
      <div className="space-y-3">
        <h2 className="font-display text-primary text-xl tracking-wide">
          All cases
        </h2>
        <RecordCaseList cases={rosterCases} canMutate={canMutate} />
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create the wizard component**

Create `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/RecordSessionWizard.tsx`:

```tsx
"use client";

import { useRouter } from "next/navigation";

import { useInfraredRecordSessionBroadcast } from "@/hooks/use-infrared-record-session-broadcast";
import { useSmartRefresh } from "@/hooks/use-smart-refresh";
import type {
  InfraredRecordSessionResponse,
  InfraredStateDeviceRecordCaseResponse,
  InfraredTestCaseResponse,
} from "@/lib/api/infrared";

import WizardStepper from "../../_components/WizardStepper";
import { getBroadcastToken } from "../_lib/broadcast-token";
import RecordCaseList from "./RecordCaseList";
import RecordTestCaseList from "./RecordTestCaseList";
import WizardCaseRecorder from "./WizardCaseRecorder";

const WIZARD_STEPS = [
  "Device details",
  "State values",
  "Building cases",
  "Review & set default",
  "Recording",
  "Analyzing",
  "Testing",
  "Finishing",
] as const;

function currentStepIndex(
  recordingState: InfraredRecordSessionResponse["recording_state"],
  anyRawExists: boolean,
): number {
  switch (recordingState) {
    case "DRAFT":
    case "CASES_GENERATING":
      return 2;
    case "RECORDING":
      return anyRawExists ? 4 : 3;
    case "ANALYZING":
    case "FUNCTION_GENERATING":
      return 5;
    case "TEST_CASES_GENERATING":
    case "TESTING":
      return 6;
    case "COMPLETED":
    case "FAILED":
      return 7;
  }
}

export default function RecordSessionWizard({
  session,
  cases,
  testCases,
  canMutate,
}: {
  session: InfraredRecordSessionResponse;
  cases: readonly InfraredStateDeviceRecordCaseResponse[];
  testCases: readonly InfraredTestCaseResponse[];
  canMutate: boolean;
}) {
  const router = useRouter();
  const { exhausted } = useInfraredRecordSessionBroadcast(
    session.id,
    () => router.refresh(),
    getBroadcastToken,
  );
  // Live updates arrive over the WebSocket; polling only takes over once
  // reconnect attempts are exhausted (e.g. a second tab already holds the
  // session's single allowed listener).
  useSmartRefresh({
    intervalMs: 5000,
    suspended: !exhausted,
    onRefresh: () => {
      router.refresh();
    },
  });

  const anyRawExists = cases.some((c) => c.raw.length > 0);
  const stepIndex = currentStepIndex(session.recording_state, anyRawExists);

  return (
    <div className="space-y-6">
      <WizardStepper steps={WIZARD_STEPS} currentIndex={stepIndex} />

      {session.recording_state === "DRAFT" ||
      session.recording_state === "CASES_GENERATING" ? (
        <div className="border-border bg-muted rounded-2xl border p-6 text-center">
          <p className="font-semibold">Building the recording plan…</p>
          <p className="text-foreground/70 mt-1 text-sm">
            The AI is drafting cases for every state combination. This page
            updates automatically.
          </p>
        </div>
      ) : null}

      {session.recording_state === "RECORDING" && !anyRawExists ? (
        <div className="border-border space-y-4 rounded-2xl border p-6">
          <div>
            <p className="font-semibold">Review the recording plan</p>
            <p className="text-foreground/70 mt-1 text-sm">
              Each case below asks you to press the remote once its target
              state is active. Before you start, set the device to its
              default/off state so the first capture has a clean baseline.
            </p>
          </div>
          <RecordCaseList cases={cases} canMutate={false} />
        </div>
      ) : null}

      {session.recording_state === "RECORDING" && anyRawExists ? (
        <WizardCaseRecorder cases={cases} canMutate={canMutate} />
      ) : null}

      {session.recording_state === "ANALYZING" ||
      session.recording_state === "FUNCTION_GENERATING" ? (
        <div className="border-border bg-muted rounded-2xl border p-6 text-center">
          <p className="font-semibold">
            {session.recording_state === "FUNCTION_GENERATING"
              ? "Refining the encoder…"
              : "Analyzing captures…"}
          </p>
          <p className="text-foreground/70 mt-1 text-sm">
            This page updates automatically once the coder is ready to test.
          </p>
        </div>
      ) : null}

      {session.recording_state === "TEST_CASES_GENERATING" ? (
        <div className="border-border bg-muted rounded-2xl border p-6 text-center">
          <p className="font-semibold">Building test cases…</p>
        </div>
      ) : null}

      {session.recording_state === "TESTING" ? (
        <RecordTestCaseList testCases={testCases} canMutate={canMutate} />
      ) : null}
    </div>
  );
}
```

- [ ] **Step 3: Branch the detail page between the static view and the wizard**

In `frontend/src/app/(authenticated)/apps/infrared/record/[id]/page.tsx`, add the import:

```tsx
import RecordSessionWizard from "./_components/RecordSessionWizard";
```

Then, immediately after `const canDelete = permissions.has("infrared_record_session:delete");` and before the existing `return (`, insert the wizard branch:

```tsx
  if (!session.is_completed) {
    return (
      <main className="mx-auto w-full max-w-5xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
        <PageHeader
          title="Record Session"
          description={`Session ${session.id}`}
        />
        <RecordSessionWizard
          session={session}
          cases={cases}
          testCases={testCases ?? []}
          canMutate={canMutate}
        />
      </main>
    );
  }

```

The existing `return (` block below it (the static overview/cases/coder/test-cases view) is now only reached for `COMPLETED`/`FAILED` sessions — leave it exactly as-is, no other edits to that file.

- [ ] **Step 4: Lint, format, typecheck, build**

Run: `cd frontend && npm run lint && npx prettier --check . && npx tsc --noEmit && npm run build`
Expected: PASS.

- [ ] **Step 5: Manual browser verification**

Using the session created in Task 4's verification (or a fresh one), open `/apps/infrared/record/{id}` while it's non-terminal:
- Confirm the stepper renders and highlights the correct step for the session's current `recording_state`.
- Confirm the page updates live (no manual refresh) as `recording_state` changes and as raw captures arrive — watch the browser's Network/WS panel to confirm a `wss://.../broadcast` connection is open.
- Accept/discard a capture, retry a case, transmit/pass/fail a test case — confirm each still works exactly as in Spec A (these are the same reused components/actions).
- Reload the tab mid-recording (before any raw exists, and again after some raws exist) — confirm the wizard lands on the same step both times, matching the session's real state.
- Once the session reaches `COMPLETED` or `FAILED`, confirm the page shows Spec A's static view (no stepper, no wizard chrome).
- Open the same session in a second browser tab — confirm the second tab still updates (via the polling fallback) even though only one tab can hold the WebSocket listener.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/'(authenticated)'/apps/infrared/record/'[id]'/_components/WizardCaseRecorder.tsx frontend/src/app/'(authenticated)'/apps/infrared/record/'[id]'/_components/RecordSessionWizard.tsx frontend/src/app/'(authenticated)'/apps/infrared/record/'[id]'/page.tsx
git commit -m "feat(infrared): add live record session wizard and route non-terminal sessions to it"
```

---

### Task 7: Frontend — restore the "Record New Device" entry point on the list page

**Files:**
- Modify: `frontend/src/app/(authenticated)/apps/infrared/record/page.tsx`

**Interfaces:**
- Consumes: Task 4's `/record/new` route, `permissions.has("infrared_record_session:add")` (already destructured in this file's `requirePermission` call).

- [ ] **Step 1: Add the header button**

In `frontend/src/app/(authenticated)/apps/infrared/record/page.tsx`, add the import:

```tsx
import Link from "next/link";
```

Change the `PageHeader` usage from:

```tsx
      <PageHeader
        title="Record"
        description="Identify the device control by capturing its remote's infrared signals."
      />
```

to:

```tsx
      <PageHeader
        title="Record"
        description="Identify the device control by capturing its remote's infrared signals."
        actions={
          permissions.has("infrared_record_session:add") ? (
            <Link
              href="/apps/infrared/record/new"
              className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background inline-flex min-h-11 w-full items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-colors hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none sm:w-auto"
            >
              Record New Device
            </Link>
          ) : null
        }
      />
```

- [ ] **Step 2: Lint, format, typecheck, build**

Run: `cd frontend && npm run lint && npx prettier --check . && npx tsc --noEmit && npm run build`
Expected: PASS.

- [ ] **Step 3: Manual browser verification**

Visit `/apps/infrared/record` as a user with `infrared_record_session:add` — confirm the "Record New Device" button appears in the header and links to `/apps/infrared/record/new`. Visit as a user without that permission — confirm the button is absent.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/'(authenticated)'/apps/infrared/record/page.tsx
git commit -m "feat(infrared): restore the Record New Device entry point on the list page"
```

---

## Final Verification

- [ ] Backend: `cd backend && go build ./... && go test ./...` — all green.
- [ ] Frontend: `cd frontend && npm run lint && npx prettier --check . && npx tsc --noEmit && npm run build` — all green.
- [ ] Full end-to-end walkthrough on `localhost` (not `127.0.0.1`): start a new session from `/apps/infrared/record/new` through to a `COMPLETED` or `FAILED` result, confirming every step's live update, every accept/discard/retry/transmit/pass-fail control, and the terminal-state handoff back to Spec A's static view.
