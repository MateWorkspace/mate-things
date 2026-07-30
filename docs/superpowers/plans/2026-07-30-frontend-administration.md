# Frontend Administration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver Users, consolidated Access Control, Payload Schemas, and contextual preference editing.

**Architecture:** Administration routes use the same server-rendered card collections and permission checks as fleet pages. Role-permission joins remain an implementation detail inside role details, while structured JSON editors stay small Client Components that submit through permission-checking Server Actions.

**Tech Stack:** Next.js 16, React 19, TypeScript, Tailwind CSS v4, Vitest, Testing Library, Playwright.

## Global Constraints

- Complete the first three plans before this phase.
- Access Control combines Roles and Permissions; do not create a separate
  role-assignment sidebar route.
- Every administrative mutation checks its exact permission server-side.
- Preferences remain contextual to their owning resource.

---

### Task 1: JSON editor and contextual preferences

**Files:**
- Create: `frontend/src/components/json/JsonEditor.tsx`
- Create: `frontend/src/components/json/JsonEditor.test.tsx`
- Create: `frontend/src/components/preferences/PreferencesDialog.tsx`
- Create: `frontend/src/components/preferences/preferences-actions.ts`
- Create: `frontend/src/components/preferences/PreferencesDialog.test.tsx`

**Interfaces:**
- `JsonEditor` consumes `name`, `label`, and `defaultValue: unknown`; submits
  formatted JSON text.
- `PreferencesDialog` consumes `resource: PreferencesResource`, `id`, current
  preferences, and permissions.

- [ ] **Step 1: Write JSON validation tests**

```tsx
it("formats valid JSON and reports invalid JSON", async () => {
  render(<JsonEditor name="definition" label="Definition" defaultValue={{ a: 1 }} />);
  await userEvent.clear(screen.getByLabelText("Definition"));
  await userEvent.type(screen.getByLabelText("Definition"), "{invalid");
  expect(screen.getByRole("alert")).toHaveTextContent("valid JSON");
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/components/json src/components/preferences
```

- [ ] **Step 3: Implement editor and action**

Use a labeled monospace textarea, Parse, Format, and Copy controls. Parsing
must accept objects and reject primitives for schema definitions/preferences.
`savePreferencesAction` checks `preferences:set`, parses the JSON object, and
calls `updatePreferences(resource, id, preferences)`.

- [ ] **Step 4: Integrate permission-aware dialog**

Show the Preferences action on Action, Firmware, Node, Node Class, Payload
Schema, Permission, Role, and User details only when `preferences:set` exists.

- [ ] **Step 5: Test and commit**

```bash
npm test -- src/components/json src/components/preferences
npm run typecheck
git add frontend/src/components/json frontend/src/components/preferences
git commit -m "feat(frontend): add structured preference editing"
```

### Task 2: User administration

**Files:**
- Create: `frontend/src/app/(authenticated)/admin/users/page.tsx`
- Create: `frontend/src/app/(authenticated)/admin/users/[id]/page.tsx`
- Create: `frontend/src/app/(authenticated)/admin/users/_components/UserCard.tsx`
- Create: `frontend/src/app/(authenticated)/admin/users/_components/UserForm.tsx`
- Create: `frontend/src/app/(authenticated)/admin/users/_components/PasswordResetForm.tsx`
- Create: `frontend/src/app/(authenticated)/admin/users/_lib/actions.ts`
- Test: `frontend/src/app/(authenticated)/admin/users/_lib/actions.test.ts`
- Test: `frontend/src/app/(authenticated)/admin/users/_components/UserCard.test.tsx`

**Interfaces:**
- Consumes: user/role/permission wrappers, collection primitives, preferences.
- Produces: user CRUD, effective permissions, role assignment, password reset.

- [ ] **Step 1: Write exact-permission tests**

```ts
it("requires user_password:set for an administrative reset", async () => {
  mockPermissions(["user:get", "user:set"]);
  const result = await resetUserPasswordAction(
    EMPTY_STATE,
    formData({ user_id: "u2", password: "new-secret" }),
  );
  expect(result.status).toBe("error");
  expect(updateUserPassword).not.toHaveBeenCalled();
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/admin/users
```

- [ ] **Step 3: Implement user cards and forms**

Cards show name, username, bio summary, resolved role, and audit context.
Create requires role/name/username/password. Edit supports role/name/bio/
username. User detail fetches effective permissions only with
`user_permission:get`.

- [ ] **Step 4: Protect dangerous account operations**

Password reset has explicit target identity. Deleting the signed-in user
requires typed username confirmation and warns that the session may end.
Backend failures remain authoritative.

- [ ] **Step 5: Test and commit**

```bash
npm test -- src/app/'(authenticated)'/admin/users
npm run typecheck
npm run lint
git add frontend/src/app/'(authenticated)'/admin/users
git commit -m "feat(frontend): add user administration"
```

### Task 3: Consolidated Access Control

**Files:**
- Create: `frontend/src/app/(authenticated)/admin/access-control/page.tsx`
- Create: `frontend/src/app/(authenticated)/admin/access-control/_components/AccessTabs.tsx`
- Create: `frontend/src/app/(authenticated)/admin/access-control/_components/RoleCard.tsx`
- Create: `frontend/src/app/(authenticated)/admin/access-control/_components/RoleDetails.tsx`
- Create: `frontend/src/app/(authenticated)/admin/access-control/_components/PermissionCard.tsx`
- Create: `frontend/src/app/(authenticated)/admin/access-control/_components/PermissionGroups.tsx`
- Create: `frontend/src/app/(authenticated)/admin/access-control/_lib/actions.ts`
- Test: `frontend/src/app/(authenticated)/admin/access-control/_components/PermissionGroups.test.tsx`
- Test: `frontend/src/app/(authenticated)/admin/access-control/_lib/actions.test.ts`

**Interfaces:**
- Consumes: roles, permissions, role-permissions, tabs, cards, exact
  `role:*`, `permission:*`, and `role_permission:*` permissions.
- Produces: `?tab=roles|permissions`, role CRUD/default/assignments, permission
  CRUD.

- [ ] **Step 1: Write grouping and read-only tests**

```tsx
it("groups permissions by product purpose", () => {
  render(<PermissionGroups permissions={PERMISSIONS} selected={new Set()} editable />);
  expect(screen.getByRole("group", { name: "Fleet" })).toBeVisible();
  expect(screen.getByRole("group", { name: "Observability" })).toBeVisible();
});

it("renders assignments without toggles when read-only", () => {
  render(<PermissionGroups permissions={PERMISSIONS} selected={SELECTED} editable={false} />);
  expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/admin/access-control
```

- [ ] **Step 3: Implement tabs and cards**

Use URL tab state. Roles show default badge, description, permission count, and
actions. Permissions show name, description, and actions. Hide a tab if the
user lacks every read permission for that resource.

- [ ] **Step 4: Implement assignment diff action**

Submit desired permission IDs. Load current IDs, then calculate:

```ts
const assign = desired.filter((id) => !current.has(id));
const revoke = [...current].filter((id) => !desiredSet.has(id));
```

Require `role_permission:add` when `assign` is non-empty and
`role_permission:remove` when `revoke` is non-empty. Execute independent
changes with `Promise.all`, report partial failure without claiming full
success, and reload authoritative assignments.

- [ ] **Step 5: Implement role/permission CRUD and default role**

Use exact permission checks for each mutation. Default-role confirmation names
the role. Permission deletion warns that assignments can cause backend
rejection.

- [ ] **Step 6: Test and commit**

```bash
npm test -- src/app/'(authenticated)'/admin/access-control
npm run typecheck
npm run lint
git add frontend/src/app/'(authenticated)'/admin/access-control
git commit -m "feat(frontend): add consolidated access control"
```

### Task 4: Versioned Payload Schemas

**Files:**
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/page.tsx`
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/[id]/page.tsx`
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/PayloadSchemaCard.tsx`
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/PayloadSchemaForm.tsx`
- Create: `frontend/src/app/(authenticated)/admin/payload-schemas/_lib/actions.ts`
- Test: `frontend/src/app/(authenticated)/admin/payload-schemas/_lib/actions.test.ts`
- Test: `frontend/src/app/(authenticated)/admin/payload-schemas/_components/PayloadSchemaCard.test.tsx`

**Interfaces:**
- Consumes: payload-schema wrappers, JsonEditor, preferences, pagination.
- Produces: version cards, validity filtering, exact/latest lookup, CRUD.

- [ ] **Step 1: Write schema parsing tests**

```ts
it("rejects a non-object definition before the API call", async () => {
  const result = await createPayloadSchemaAction(
    EMPTY_STATE,
    formData({ name: "sensor", version: "2", definition: "[]", valid_from: "" }),
  );
  expect(result.fieldErrors?.definition).toMatch(/object/i);
  expect(createPayloadSchema).not.toHaveBeenCalled();
});
```

- [ ] **Step 2: Verify failure**

```bash
npm test -- src/app/'(authenticated)'/admin/payload-schemas
```

- [ ] **Step 3: Implement versioned cards and filters**

Cards show name/version, valid-from/to, current/expired/future status, and audit
context. URL filters support search and `valid_at`. Details provide formatted
definition and references back to filtered Actions.

- [ ] **Step 4: Implement form actions**

Parse integer version, JSON object definition, and ISO validity dates. Require
`payload_schema:add|set|remove` as appropriate. Preserve backend validation
messages when date/version conflicts occur.

- [ ] **Step 5: Verify phase and commit**

```bash
npm test
npm run typecheck
npm run lint
npm run build
git add frontend/src/app/'(authenticated)'/admin/payload-schemas
git commit -m "feat(frontend): add payload schema management"
```
