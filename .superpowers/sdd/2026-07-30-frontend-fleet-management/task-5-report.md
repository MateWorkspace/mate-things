# Task 5 Implementation Report: Firmware Binaries and OTA

## Result

Implemented the firmware collection and detail workflows plus node-scoped OTA
dispatch on `feature/frontend-operations-cockpit`.

Planned commit:

- `feat(frontend): add firmware and OTA workflows`

## Files

### API and runtime configuration

- `frontend/src/lib/api/firmwares.ts`
  - sends JSON-encoded `config_schema` entries with create and binary-replace
    multipart requests;
  - exports the shared `FirmwareConfigSchemaItem` contract.
- `frontend/src/lib/api/firmwares.test.ts`
  - verifies multipart metadata, file, and config-schema payloads.
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

## Verification

- Focused Task 5 suite: 8 test files, 34 tests passed.
- Full frontend suite: 44 test files, 155 tests passed.
- `npm run typecheck`: passed.
- `npm run lint`: passed.
- Task 5 scoped Prettier checks: passed.
- `npm run build`: passed with Next.js 16.2.11/Turbopack; `/firmware`,
  `/firmware/[id]`, and `/nodes/[id]` build as dynamic routes.
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
- The current backend preserves an existing configuration schema when binary
  replacement submits an empty array (`ReplaceBinaryById` only replaces
  parameters when `len(config_schema) > 0`). The frontend sends the exact
  submitted array and leaves that rule authoritative to the backend.
- No internal reviewer was run; the controller owns the mandatory review.
