# Frontend Foundation, Shell, and Account Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the tested UI foundation, effective-permission session model, responsive authenticated shell, protected routing, and editable profile dialog.

**Architecture:** Server Components resolve the profile and permissions once per request and pass plain values into a small client-side shell. Pure permission/navigation helpers remain unit-testable, while profile mutations use Server Actions and existing server-only API wrappers.

**Tech Stack:** Next.js 16.2.11, React 19.2.4, TypeScript, Tailwind CSS v4, Lucide React, Vitest, Testing Library, Playwright.

## Global Constraints

- Follow all constraints in `2026-07-30-frontend-operations-cockpit.md`.
- Read the local Next.js guides for layouts/pages, Server/Client Components,
  authentication, forms, route groups, Vitest, and Playwright before editing.
- Do not add a state-management library or client-side API cache.
- Keep the existing `/login` behavior and same-origin production API topology.

---

### Task 1: Testing and verification harness

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/package-lock.json`
- Create: `frontend/vitest.config.mts`
- Create: `frontend/vitest.setup.ts`
- Create: `frontend/playwright.config.ts`
- Create: `frontend/src/components/ui/button.test.tsx`
- Create: `frontend/e2e/login.spec.ts`
- Create: `frontend/src/test/form-data.ts`
- Create: `frontend/src/test/fixtures.ts`

**Interfaces:**
- Consumes: existing `Button`, `/login`, and environment configuration.
- Produces: `npm test`, `npm run test:watch`, `npm run typecheck`, and
  `npm run test:e2e`.

- [ ] **Step 1: Add scripts and test dependencies**

Add scripts:

```json
{
  "test": "vitest run",
  "test:watch": "vitest",
  "typecheck": "tsc --noEmit",
  "test:e2e": "playwright test"
}
```

Install exact current-compatible packages through npm so the lockfile records
resolved versions:

```bash
npm install --save-dev vitest @vitejs/plugin-react jsdom @testing-library/react @testing-library/jest-dom @testing-library/user-event @playwright/test @axe-core/playwright
```

- [ ] **Step 2: Configure Vitest**

Create:

```ts
// vitest.config.mts
import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    setupFiles: ["./vitest.setup.ts"],
    include: ["src/**/*.test.{ts,tsx}"],
  },
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
  },
});
```

```ts
// vitest.setup.ts
import "@testing-library/jest-dom/vitest";

HTMLDialogElement.prototype.showModal = function showModal() {
  this.open = true;
};
HTMLDialogElement.prototype.close = function close() {
  this.open = false;
};
```

- [ ] **Step 3: Add shared deterministic test data**

```ts
// src/test/form-data.ts
export function formData(
  values: Record<string, string | Blob>,
): FormData {
  const data = new FormData();
  for (const [key, value] of Object.entries(values)) data.set(key, value);
  return data;
}
```

```ts
// src/test/fixtures.ts
import type { NodeResponse, UserResponse } from "@/lib/api";

export const USER: UserResponse = {
  id: "user-1",
  role_id: "role-1",
  name: "Alex Morgan",
  bio: "Fleet operator",
  username: "alex",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

export function nodeFixture(
  overrides: Partial<NodeResponse> = {},
): NodeResponse {
  return {
    id: "node-1",
    node_class_id: "class-1",
    device_id: "AC276E5E030C",
    device_info: "ESP32",
    name: "Cold Storage Sensor 07",
    firmware_id: "firmware-1",
    description: "Freezer room sensor",
    is_connected: false,
    preferences: {},
    created_at: "2026-07-30T00:00:00Z",
    ...overrides,
  };
}

export const EMPTY_STATE = { status: "idle" } as const;
```

- [ ] **Step 4: Write the first component test**

```tsx
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Button from "./button";

describe("Button", () => {
  it("preserves native accessibility and applies its variant", () => {
    render(<Button variant="secondary">Cancel</Button>);
    expect(screen.getByRole("button", { name: "Cancel" })).toHaveClass(
      "text-primary",
    );
  });
});
```

- [ ] **Step 5: Configure Playwright and smoke-test login**

```ts
// playwright.config.ts
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  use: {
    baseURL: "http://127.0.0.1:3000",
    trace: "on-first-retry",
  },
  webServer: {
    command: "npm run dev",
    url: "http://127.0.0.1:3000/login",
    reuseExistingServer: !process.env.CI,
  },
  projects: [
    { name: "chromium", use: { ...devices["Desktop Chrome"] } },
  ],
});
```

```ts
// e2e/login.spec.ts
import { expect, test } from "@playwright/test";

test("login page exposes an accessible credential form", async ({ page }) => {
  await page.goto("/login");
  await expect(page.getByRole("heading", { name: "Sign in" })).toBeVisible();
  await expect(page.getByLabel("Username")).toBeVisible();
  await expect(page.getByLabel("Password")).toHaveAttribute("type", "password");
});
```

- [ ] **Step 6: Run verification**

Run:

```bash
npm test
npm run typecheck
npm run lint
```

Expected: all commands exit 0.

- [ ] **Step 7: Commit**

```bash
git add frontend/package.json frontend/package-lock.json frontend/vitest.config.mts frontend/vitest.setup.ts frontend/playwright.config.ts frontend/src/test frontend/src/components/ui/button.test.tsx frontend/e2e/login.spec.ts
git commit -m "test(frontend): add component and e2e harness"
```

### Task 2: Semantic theme and reusable primitives

**Files:**
- Modify: `frontend/src/app/globals.css`
- Modify: `frontend/src/components/ui/button.tsx`
- Modify: `frontend/src/components/ui/input.tsx`
- Create: `frontend/src/components/ui/icon-button.tsx`
- Create: `frontend/src/components/ui/card.tsx`
- Create: `frontend/src/components/ui/status-badge.tsx`
- Create: `frontend/src/components/ui/dialog.tsx`
- Create: `frontend/src/components/ui/tabs.tsx`
- Create: `frontend/src/components/ui/states.tsx`
- Test: `frontend/src/components/ui/status-badge.test.tsx`
- Test: `frontend/src/components/ui/dialog.test.tsx`

**Interfaces:**
- Produces: shared primitives used by every later route.
- `StatusBadge` consumes `variant: "success" | "warning" | "critical" |
  "info" | "neutral"` and visible children.
- `Dialog` consumes `open`, `onClose`, `title`, and `children`.

- [ ] **Step 1: Write failing status and dialog tests**

```tsx
it("pairs critical color with visible text", () => {
  render(<StatusBadge variant="critical">Disconnected</StatusBadge>);
  expect(screen.getByText("Disconnected")).toHaveAttribute(
    "data-variant",
    "critical",
  );
});

it("labels the modal and closes on Escape", async () => {
  const onClose = vi.fn();
  render(
    <Dialog open onClose={onClose} title="Edit profile">
      Content
    </Dialog>,
  );
  await userEvent.keyboard("{Escape}");
  expect(onClose).toHaveBeenCalledOnce();
});
```

- [ ] **Step 2: Verify the tests fail**

Run:

```bash
npm test -- src/components/ui/status-badge.test.tsx src/components/ui/dialog.test.tsx
```

Expected: fail because the components do not exist.

- [ ] **Step 3: Add two-mode semantic tokens**

Extend `:root`, dark media values, and `@theme inline` with:

```css
--muted: #f8f2e8;
--border: color-mix(in srgb, #231f20 14%, transparent);
--success: #287a52;
--warning: #9a5b13;
--critical: #b43b35;
--info: #356f91;
--destructive: #a8342f;
--focus: #8c6038;
```

Use darker-surface, lighter-foreground equivalents in the existing dark media
block. Add a reduced-motion rule that disables nonessential transition and
animation durations.

- [ ] **Step 4: Implement focused primitives**

`StatusBadge` must render a visible icon and label:

```tsx
const STATUS_ICON = {
  success: CircleCheck,
  warning: TriangleAlert,
  critical: CircleX,
  info: Info,
  neutral: Circle,
} satisfies Record<StatusVariant, LucideIcon>;

export default function StatusBadge({ variant, children }: StatusBadgeProps) {
  const Icon = STATUS_ICON[variant];
  return (
    <span data-variant={variant} className={VARIANT_CLASSES[variant]}>
      <Icon aria-hidden="true" className="size-3.5" />
      {children}
    </span>
  );
}
```

`Dialog` must use the native modal dialog API, call `showModal()` when opened,
close on cancel, restore focus, and provide `aria-labelledby`.

- [ ] **Step 5: Run primitive tests and visual static checks**

Run:

```bash
npm test -- src/components/ui
npm run typecheck
npm run lint
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/globals.css frontend/src/components/ui
git commit -m "feat(frontend): add operations UI primitives"
```

### Task 3: Effective permissions and navigation model

**Files:**
- Create: `frontend/src/lib/permissions.ts`
- Create: `frontend/src/lib/permissions.test.ts`
- Create: `frontend/src/config/navigation.ts`
- Create: `frontend/src/config/navigation.test.ts`
- Modify: `frontend/src/lib/session/session.ts`
- Modify: `frontend/src/lib/session/index.ts`

**Interfaces:**
- Produces: `PermissionName`, `SessionContext`, `hasPermission`,
  `canAccessAny`, `getSessionContext`, and `NAVIGATION_GROUPS`.
- Consumes: `getProfile()`, `getProfilePermissions()`.

- [ ] **Step 1: Write failing permission tests**

```ts
it("omits empty navigation groups", () => {
  const visible = visibleNavigation(new Set(["node:get"]));
  expect(visible.map((group) => group.label)).toEqual(["Overview", "Fleet"]);
  expect(visible[1].items.map((item) => item.label)).toEqual(["Nodes"]);
});

it("shows Access Control for any access-control read permission", () => {
  const visible = visibleNavigation(new Set(["role:get"]));
  expect(visible.flatMap((group) => group.items).map((item) => item.label))
    .toContain("Access Control");
});
```

- [ ] **Step 2: Verify failure**

Run:

```bash
npm test -- src/lib/permissions.test.ts src/config/navigation.test.ts
```

Expected: fail because helpers do not exist.

- [ ] **Step 3: Implement the pure contracts**

```ts
export type PermissionName = string;

export function hasPermission(
  permissions: ReadonlySet<string>,
  permission: string,
): boolean {
  return permissions.has(permission);
}

export function canAccessAny(
  permissions: ReadonlySet<string>,
  required: readonly string[],
): boolean {
  return required.length === 0 || required.some((item) => permissions.has(item));
}
```

Define navigation data with exact routes and read permissions from the design
spec. `visibleNavigation()` filters items first and groups second.

- [ ] **Step 4: Add the cached server session context**

```ts
export interface SessionContext {
  user: UserResponse;
  permissions: ReadonlySet<PermissionName>;
}

export const getSessionContext = cache(
  async (): Promise<SessionContext | null> => {
    const user = await getSession();
    if (!user) return null;
    const permissions = await getProfilePermissions();
    return {
      user,
      permissions: new Set(permissions.map((permission) => permission.name)),
    };
  },
);
```

Add `requireSessionContext()` which redirects to `/login` when absent.

- [ ] **Step 5: Run tests**

Run:

```bash
npm test -- src/lib/permissions.test.ts src/config/navigation.test.ts
npm run typecheck
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/permissions.ts frontend/src/lib/permissions.test.ts frontend/src/config/navigation.ts frontend/src/config/navigation.test.ts frontend/src/lib/session
git commit -m "feat(frontend): model effective navigation permissions"
```

### Task 4: Responsive authenticated shell

**Files:**
- Create: `frontend/src/app/(authenticated)/layout.tsx`
- Move: `frontend/src/app/dashboard/page.tsx` to `frontend/src/app/(authenticated)/dashboard/page.tsx`
- Move: `frontend/src/app/dashboard/_components/LogoutButton.tsx` to `frontend/src/components/layout/LogoutButton.tsx`
- Move: `frontend/src/app/dashboard/_lib/actions.ts` to `frontend/src/lib/api/session-actions.ts`
- Create: `frontend/src/components/layout/AppShell.tsx`
- Create: `frontend/src/components/layout/AppBar.tsx`
- Create: `frontend/src/components/layout/Sidebar.tsx`
- Create: `frontend/src/components/layout/SidebarGroup.tsx`
- Create: `frontend/src/components/layout/AppShell.test.tsx`

**Interfaces:**
- `AppShell` consumes `user: UserResponse`, `permissions: readonly string[]`,
  and `children`.
- `Sidebar` consumes filtered navigation and current pathname.
- Produces the shared layout for every authenticated route.

- [ ] **Step 1: Write the failing shell test**

```tsx
it("toggles labels on desktop and keeps permitted links", async () => {
  render(
    <AppShell user={USER} permissions={["node:get"]}>
      <main>Page</main>
    </AppShell>,
  );
  expect(screen.getByRole("link", { name: "Nodes" })).toBeVisible();
  expect(screen.queryByRole("link", { name: "Users" })).not.toBeInTheDocument();
  await userEvent.click(screen.getByRole("button", { name: "Collapse sidebar" }));
  expect(screen.getByRole("button", { name: "Expand sidebar" })).toBeVisible();
});
```

- [ ] **Step 2: Verify failure**

Run:

```bash
npm test -- src/components/layout/AppShell.test.tsx
```

Expected: fail because the shell does not exist.

- [ ] **Step 3: Implement the Server Component layout**

```tsx
export default async function AuthenticatedLayout({ children }: LayoutProps) {
  const session = await requireSessionContext();
  return (
    <AppShell
      user={session.user}
      permissions={[...session.permissions]}
    >
      {children}
    </AppShell>
  );
}
```

- [ ] **Step 4: Implement the client shell**

`AppShell` owns only sidebar/drawer state. `AppBar` receives toggle callbacks
and profile data. `Sidebar` renders `visibleNavigation()` and uses `usePathname`
for `aria-current`. Persist desktop collapsed state in a non-sensitive cookie
through a tiny Server Action; do not use local storage.

- [ ] **Step 5: Verify shell behavior**

Run:

```bash
npm test -- src/components/layout/AppShell.test.tsx
npm run typecheck
npm run lint
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/'(authenticated)' frontend/src/app/dashboard frontend/src/components/layout frontend/src/lib/api/session-actions.ts
git commit -m "feat(frontend): add permission-aware app shell"
```

### Task 5: Editable profile and security dialog

**Files:**
- Create: `frontend/src/components/profile/ProfileDialog.tsx`
- Create: `frontend/src/components/profile/ProfileView.tsx`
- Create: `frontend/src/components/profile/ProfileForm.tsx`
- Create: `frontend/src/components/profile/SecurityForm.tsx`
- Create: `frontend/src/components/profile/profile-actions.ts`
- Create: `frontend/src/components/profile/ProfileDialog.test.tsx`
- Modify: `frontend/src/components/layout/AppBar.tsx`

**Interfaces:**
- Consumes: `UserResponse`, effective permission names, `updateProfile`,
  `updateProfilePassword`, and `logout`.
- Produces: profile trigger/dialog used by `AppBar`.

- [ ] **Step 1: Write failing dialog-flow tests**

```tsx
it("opens from the profile trigger and enters edit mode", async () => {
  render(<AppBar user={USER} permissions={["profile:get", "profile:set"]} />);
  await userEvent.click(screen.getByRole("button", { name: /Alex Morgan/ }));
  expect(screen.getByRole("dialog", { name: "Your profile" })).toBeVisible();
  await userEvent.click(screen.getByRole("button", { name: "Edit profile" }));
  expect(screen.getByLabelText("Name")).toHaveValue("Alex Morgan");
});

it("omits edit controls without profile:set", async () => {
  render(<ProfileDialog open user={USER} permissions={["profile:get"]} />);
  expect(screen.queryByRole("button", { name: "Edit profile" }))
    .not.toBeInTheDocument();
});
```

- [ ] **Step 2: Verify failure**

Run:

```bash
npm test -- src/components/profile/ProfileDialog.test.tsx
```

Expected: fail because profile components do not exist.

- [ ] **Step 3: Implement typed Server Actions**

Each action returns:

```ts
export interface FormActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
}
```

`saveProfileAction` trims name/username/bio, requires non-empty name and
username, checks `profile:set`, calls `updateProfile`, and refreshes the current
route. `changePasswordAction` checks confirmation equality and
`profile_security:set` before calling the backend.

- [ ] **Step 4: Implement Profile and Security tabs**

The dialog opens in view mode. Profile edit uses `useActionState`; Security
contains current/new/confirm password fields. Keep logout separated in the
dialog footer. Announce action results through inline text and toast.

- [ ] **Step 5: Run tests**

Run:

```bash
npm test -- src/components/profile
npm run typecheck
npm run lint
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/profile frontend/src/components/layout/AppBar.tsx
git commit -m "feat(frontend): add editable account dialog"
```

### Task 6: Protected routing and foundation dashboard

**Files:**
- Modify: `frontend/src/proxy.ts`
- Modify: `frontend/src/app/(authenticated)/dashboard/page.tsx`
- Create: `frontend/src/app/(authenticated)/dashboard/loading.tsx`
- Create: `frontend/src/app/(authenticated)/dashboard/error.tsx`
- Create: `frontend/src/components/ui/page-header.tsx`
- Create: `frontend/src/proxy.test.ts`

**Interfaces:**
- Consumes: the authenticated route group and AppShell.
- Produces: protection for every signed-in route prefix and a real dashboard
  frame ready for aggregate data in phase 5.

- [ ] **Step 1: Write route classification tests**

```ts
it.each([
  "/dashboard",
  "/nodes",
  "/firmware/abc",
  "/actions",
  "/telemetry",
  "/node-logs",
  "/admin/users",
])("treats %s as protected", (pathname) => {
  expect(isProtectedRoute(pathname)).toBe(true);
});

it("does not protect login", () => {
  expect(isProtectedRoute("/login")).toBe(false);
});
```

- [ ] **Step 2: Verify failure**

Run:

```bash
npm test -- src/proxy.test.ts
```

Expected: fail for routes outside `/dashboard`.

- [ ] **Step 3: Expand protected prefixes**

Export the pure route classifier for tests and define:

```ts
const PROTECTED_PREFIXES = [
  "/dashboard",
  "/nodes",
  "/node-classes",
  "/firmware",
  "/actions",
  "/action-history",
  "/telemetry",
  "/node-logs",
  "/admin",
] as const;
```

- [ ] **Step 4: Replace the placeholder dashboard**

Render `PageHeader`, four skeleton/permission-aware metric card positions, an
urgent-attention section, and recent-warning section using honest empty copy.
Do not invent metrics before phase 5 connects data.

- [ ] **Step 5: Verify the complete phase**

Run:

```bash
npm test
npm run typecheck
npm run lint
npm run build
```

Expected: all exit 0.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/proxy.ts frontend/src/proxy.test.ts frontend/src/app/'(authenticated)'/dashboard frontend/src/components/ui/page-header.tsx
git commit -m "feat(frontend): protect operations routes"
```
