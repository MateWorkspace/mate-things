# Task 5 Implementation Report: Firmware Binaries and OTA

## Result

Implemented the firmware collection and detail workflows plus node-scoped OTA
dispatch on `feature/frontend-operations-cockpit`.

Planned commit:

- `feat(frontend): add firmware and OTA workflows`
- `fix(frontend): harden firmware and OTA workflows` (review round 1)

## Files

### API and runtime configuration

- `frontend/src/lib/api/firmwares.ts`
  - sends JSON-encoded `config_schema` entries with create and binary-replace
    multipart requests;
  - exports the shared `FirmwareConfigSchemaItem` contract.
- `frontend/src/lib/api/firmwares.test.ts`
  - verifies multipart metadata, file, empty-schema, and authoritative delete
    confirmation payloads.
- `frontend/src/lib/api/node-classes.ts`
- `frontend/src/lib/api/node-classes.test.ts`
  - loads every node-class page for firmware selectors instead of silently
    capping the available choices at 48.
- `frontend/next.config.ts`
  - raises the Next.js Server Action request limit to `8mb` so ESP32 binary
    uploads are not rejected by the framework's 1 MB default.

### Firmware collection and details

- `frontend/src/app/(authenticated)/firmware/page.tsx`
- `frontend/src/app/(authenticated)/firmware/page.test.tsx`
- `frontend/src/app/(authenticated)/firmware/[id]/page.tsx`
- `frontend/src/app/(authenticated)/firmware/[id]/page.test.tsx`
- `frontend/src/app/(authenticated)/firmware/_components/FirmwareCard.tsx`
- `frontend/src/app/(authenticated)/firmware/_components/FirmwareCard.test.tsx`
- `frontend/src/app/(authenticated)/firmware/_components/FirmwareForm.tsx`
- `frontend/src/app/(authenticated)/firmware/_components/FirmwareForm.test.tsx`
- `frontend/src/app/(authenticated)/firmware/_lib/format.ts`

These files provide card-first URL pagination and filtering, upload and metadata
editing, binary replacement, typed-name deletion, backend-resolved download
links, configuration-schema editing, binary/checksum/class/preferences/audit
details, and related-node links.

The review follow-up also adds collection/detail `loading.tsx` and `error.tsx`
boundaries, a detail `not-found.tsx`, reusable Retry states, complete
node-class options, and a fallback that keeps a firmware's current class
selected even if it is no longer returned by the active-class collection.

### Server Actions and node-scoped OTA

- `frontend/src/app/(authenticated)/firmware/_lib/actions.ts`
- `frontend/src/app/(authenticated)/firmware/_lib/actions.test.ts`
- `frontend/src/app/(authenticated)/firmware/_components/OtaDialog.tsx`
- `frontend/src/app/(authenticated)/firmware/_components/OtaDialog.test.tsx`
- `frontend/src/app/(authenticated)/nodes/[id]/_components/NodeFirmwareWorkspace.tsx`
- `frontend/src/app/(authenticated)/nodes/[id]/page.tsx`
- `frontend/src/app/(authenticated)/nodes/[id]/page.test.tsx`

Every mutation rechecks its exact permission inside the Server Action. OTA is
shown only with `ota:dispatch`, requires explicit node/firmware confirmation,
revalidates compatibility by paging only through
`listAvailableFirmwaresByNodeId`, resolves the backend binary URL server-side,
and dispatches only after the selected firmware is still available.
Successful dispatch is terminal for the open dialog: it closes, clears the
selection and confirmation, and requires a fresh confirmation when reopened.

### Backend confirmation and schema-clear contract

- `backend/internal/domain/usecases/node/firmware_management.go`
- `backend/internal/application/node/firmware_management/usecase.go`
- `backend/internal/application/node/firmware_management/usecase_test.go`
- `backend/internal/presentation/http/request/node.go`
- `backend/internal/presentation/http/handler/node/handler.go`
- `backend/internal/presentation/http/handler/node/handler_test.go`
- `backend/docs/swagger/*`

Firmware deletion now accepts `expected_name`, validates it in the application
layer, reads the authoritative firmware by ID, and refuses to delete the
database row or binary when the names differ. The route continues to require
only `firmware:remove`, so remove-only roles do not acquire a new read
dependency. Binary replacement now always forwards the submitted schema to the
existing replace repository; an explicit empty array therefore soft-deletes
all old configuration parameters. HTTP, use-case, frontend API, and Server
Action regressions cover these paths.

## Verification

- Focused review suite: 8 test files, 39 tests passed.
- Full frontend suite: 46 test files, 168 tests passed.
- `npm run typecheck`: passed.
- `npm run lint`: passed.
- Task 5 scoped Prettier write/check: passed.
- `npm run build`: passed with Next.js 16.2.11/Turbopack; `/firmware`,
  `/firmware/[id]`, and `/nodes/[id]` build as dynamic routes.
- `go test -count=1 ./...`: passed.
- `go build ./...`: passed.
- `go vet ./...`: passed.
- Swagger generation: passed; firmware DELETE documents the required
  confirmation body.
- `gofmt -l` for backend Go files: passed with no output.
- `git diff --check`: passed.

All npm commands used `PATH=/tmp/mate-node-v24.18.1/bin:$PATH`.

## Concerns

- Repository-wide `npm run format:check` still reports three pre-existing,
  unrelated files: `frontend/.superdesign/init/routes.md`,
  `frontend/.superdesign/init/theme.md`, and `frontend/AGENTS.md`. They were
  intentionally preserved; every Task 5 file passes Prettier.
- The `8mb` Server Action cap is deliberately bounded. A backend-supported
  binary larger than that would require coordinated ingress and frontend limit
  changes.
- No additional internal reviewer was run; this commit addresses the
  controller-owned review round 1 findings directly.
