# Toast Notifications Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a reusable success/error toast notification system (icon + title + message + auto-dismiss progress bar + swipe animations), replace the frontend's existing inline error text with it, and align the frontend's `ApiError` shape with the backend's current `{error, message, details}` response contract so toasts show real, curated backend copy.

**Architecture:** A presentational `Toast` component (icon/title/message/progress bar/close button, own enter/exit CSS-transition animation) rendered by a `ToastProvider` (React Context holding the active-toast queue, top-right fixed stack, 5s auto-dismiss) exposed to the rest of the app via a `useToast()` hook. `lib/api/client.ts`'s `ApiError`/`ErrorResponse` are updated to match the backend's `{error, message, details}` shape (previously stale `{code, message}`). Two real call sites are wired: `LoginForm` (error toast, replacing its inline `<p role="alert">`) and logout→login (success toast via a one-shot query param).

**Tech Stack:** Next.js 16 / React 19 (existing), `lucide-react` (new dependency) for icons, Tailwind CSS v4 for styling/animation (no new animation library).

## Global Constraints

- Icon library: `lucide-react`.
- Toast position: fixed top-right stack, newest appended below existing ones.
- Auto-dismiss: 5000ms, synced with the progress bar's width transition.
- Colors: success = green (existing brand-adjacent green, not a new hue family per AGENTS.md's "everything should read as part of the same warm, brownish-pastel family" — greens/reds are the accepted exception for semantic status color, matching how `text-red-700` is already used in the codebase for the current inline error).
- Enter animation: slide down from above + fade in. Exit animation: slide up + fade out.
- This project has no test runner configured (`package.json` has no `test` script) and is not a git repository — verification is `tsc --noEmit` + `eslint` + manual browser check (per this repo's established pattern, see the frontend's dev-server + chromium-cli-style verification used earlier in this project), and there are no git-commit steps in this plan.
- Follow `frontend/AGENTS.md`: Server Components by default, `'use client'` pushed to the leaf, mobile-first, dark/light mode both checked, semantic HTML + keyboard reachability + visible focus states.

---

### Task 1: Add lucide-react and align ApiError with the backend's current error shape

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/src/lib/api/types.ts`
- Modify: `frontend/src/lib/api/client.ts`

**Interfaces:**
- Produces: `ApiError` now has `readonly title: string`, `readonly message: string` (inherited from `Error`), `readonly details?: string`, `readonly status: number`. The old `readonly code: string` field is removed — nothing in the codebase reads `.code` today (verified: only `client.ts` itself constructs it).
- Produces: `ErrorResponse` interface now has `error`, `message`, `details` string fields (matches `backend/internal/presentation/http/response/common.go`'s current `ErrorResponse`).

- [ ] **Step 1: Add the dependency**

```bash
cd frontend && npm install lucide-react
```

- [ ] **Step 2: Update `ErrorResponse` in `frontend/src/lib/api/types.ts`**

Replace:
```ts
export interface ErrorResponse {
  code: string;
  message: string;
}
```
with:
```ts
export interface ErrorResponse {
  error: string;
  message: string;
  details: string;
}
```

- [ ] **Step 3: Update `ApiError` in `frontend/src/lib/api/client.ts`**

Replace:
```ts
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}
```
with:
```ts
export class ApiError extends Error {
  readonly status: number;
  readonly title: string;
  readonly details?: string;

  constructor(status: number, title: string, message: string, details?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.title = title;
    this.details = details;
  }
}
```

- [ ] **Step 4: Update the two `ApiError` construction sites in the same file**

Replace:
```ts
    throw new ApiError(
      response.status,
      "unexpected_redirect",
      `Unexpected redirect response for ${path} - use apiRequest directly for redirect-returning endpoints.`,
    );
```
with:
```ts
    throw new ApiError(
      response.status,
      "Unexpected Redirect",
      `Unexpected redirect response for ${path} - use apiRequest directly for redirect-returning endpoints.`,
    );
```

Replace:
```ts
    throw new ApiError(
      response.status,
      errorBody?.code ?? "unknown_error",
      errorBody?.message ?? response.statusText,
    );
```
with:
```ts
    throw new ApiError(
      response.status,
      errorBody?.error ?? "Something Went Wrong",
      errorBody?.message ?? response.statusText,
      errorBody?.details,
    );
```

- [ ] **Step 5: Verify**

```bash
cd frontend && npx tsc --noEmit && npx eslint src/lib/api/client.ts src/lib/api/types.ts
```
Expected: no output (clean). This will show type errors in any file still using `.code` — grep first to be sure:
```bash
grep -rn "\.code\b" frontend/src/lib/api/*.ts frontend/src/app
```
Expected: no remaining matches outside this task's own files.

---

### Task 2: Build the presentational Toast component

**Files:**
- Create: `frontend/src/components/ui/toast.tsx`

**Interfaces:**
- Consumes: nothing from other tasks (pure component).
- Produces: `export type ToastVariant = "success" | "error";` and `export default function Toast(props: ToastProps): JSX.Element` where
  ```ts
  interface ToastProps {
    variant: ToastVariant;
    title: string;
    message: string;
    leaving: boolean;
    onClose: () => void;
  }
  ```
  Task 3 renders one `<Toast>` per active item and passes `leaving` (true once dismissal has started, driving the exit animation) and `onClose` (calls the provider's manual-dismiss path).

- [ ] **Step 1: Write the component**

```tsx
"use client";

import { CheckCircle2, X, XCircle } from "lucide-react";
import { useEffect, useState } from "react";

export type ToastVariant = "success" | "error";

interface ToastProps {
  variant: ToastVariant;
  title: string;
  message: string;
  leaving: boolean;
  onClose: () => void;
}

const DURATION_MS = 5000;

const VARIANT_STYLES: Record<
  ToastVariant,
  { icon: typeof CheckCircle2; iconClass: string; barClass: string }
> = {
  success: {
    icon: CheckCircle2,
    iconClass: "text-emerald-600",
    barClass: "bg-emerald-600",
  },
  error: {
    icon: XCircle,
    iconClass: "text-red-600",
    barClass: "bg-red-600",
  },
};

export default function Toast({
  variant,
  title,
  message,
  leaving,
  onClose,
}: ToastProps) {
  // Starts off-screen/invisible, then flips true on the next frame so the
  // transition to the "settled" position actually animates (slide down).
  const [entered, setEntered] = useState(false);
  const [barFilled, setBarFilled] = useState(false);

  useEffect(() => {
    const raf = requestAnimationFrame(() => {
      setEntered(true);
      setBarFilled(true);
    });
    return () => cancelAnimationFrame(raf);
  }, []);

  const { icon: Icon, iconClass, barClass } = VARIANT_STYLES[variant];
  const visible = entered && !leaving;

  return (
    <div
      role="alert"
      className={`pointer-events-auto relative w-full max-w-sm overflow-hidden rounded-2xl border border-ink/10 bg-background shadow-lg transition-all duration-300 ease-out ${
        visible
          ? "translate-y-0 opacity-100"
          : "-translate-y-4 opacity-0"
      }`}
    >
      <div className="flex items-start gap-3 p-4 pr-10">
        <Icon className={`mt-0.5 h-5 w-5 shrink-0 ${iconClass}`} aria-hidden="true" />
        <div className="min-w-0 flex-1">
          <p className="font-display text-base tracking-wide text-foreground">
            {title}
          </p>
          <p className="mt-0.5 text-sm text-foreground/70">{message}</p>
        </div>
      </div>
      <button
        type="button"
        onClick={onClose}
        aria-label="Dismiss notification"
        className="absolute right-2 top-2 rounded-full p-1.5 text-foreground/50 transition-colors hover:bg-ink/10 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
      >
        <X className="h-4 w-4" aria-hidden="true" />
      </button>
      <div className="h-1 w-full bg-ink/10">
        <div
          className={`h-full ${barClass} transition-[width] ease-linear`}
          style={{
            width: barFilled ? "100%" : "0%",
            transitionDuration: `${DURATION_MS}ms`,
          }}
        />
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify types/lint**

```bash
cd frontend && npx tsc --noEmit && npx eslint src/components/ui/toast.tsx
```
Expected: clean (no output).

---

### Task 3: Build the ToastProvider (queue, timing, viewport)

**Files:**
- Create: `frontend/src/components/ui/toast-provider.tsx`

**Interfaces:**
- Consumes: `Toast` component from Task 2 (`import Toast, { type ToastVariant } from "@/components/ui/toast"`).
- Produces:
  ```ts
  export interface ToastContextValue {
    success: (title: string, message: string) => void;
    error: (title: string, message: string) => void;
  }
  export const ToastContext: React.Context<ToastContextValue | null>;
  export default function ToastProvider(props: { children: React.ReactNode }): JSX.Element;
  ```
  Task 4's `useToast()` hook consumes `ToastContext`.

- [ ] **Step 1: Write the provider**

```tsx
"use client";

import { createContext, useCallback, useRef, useState } from "react";

import Toast, { type ToastVariant } from "@/components/ui/toast";

interface ToastItem {
  id: number;
  variant: ToastVariant;
  title: string;
  message: string;
  leaving: boolean;
}

export interface ToastContextValue {
  success: (title: string, message: string) => void;
  error: (title: string, message: string) => void;
}

export const ToastContext = createContext<ToastContextValue | null>(null);

const AUTO_DISMISS_MS = 5000;
// Must match toast.tsx's exit transition duration (duration-300) so the
// item is removed from the list only after its slide-up animation finishes.
const EXIT_ANIMATION_MS = 300;

export default function ToastProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const nextId = useRef(0);

  const dismiss = useCallback((id: number) => {
    setToasts((current) =>
      current.map((toast) =>
        toast.id === id ? { ...toast, leaving: true } : toast,
      ),
    );
    setTimeout(() => {
      setToasts((current) => current.filter((toast) => toast.id !== id));
    }, EXIT_ANIMATION_MS);
  }, []);

  const push = useCallback(
    (variant: ToastVariant, title: string, message: string) => {
      const id = nextId.current++;
      setToasts((current) => [
        ...current,
        { id, variant, title, message, leaving: false },
      ]);
      setTimeout(() => dismiss(id), AUTO_DISMISS_MS);
    },
    [dismiss],
  );

  const value: ToastContextValue = {
    success: (title, message) => push("success", title, message),
    error: (title, message) => push("error", title, message),
  };

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="pointer-events-none fixed right-4 top-4 z-50 flex w-full max-w-sm flex-col gap-2">
        {toasts.map((toast) => (
          <Toast
            key={toast.id}
            variant={toast.variant}
            title={toast.title}
            message={toast.message}
            leaving={toast.leaving}
            onClose={() => dismiss(toast.id)}
          />
        ))}
      </div>
    </ToastContext.Provider>
  );
}
```

- [ ] **Step 2: Verify types/lint**

```bash
cd frontend && npx tsc --noEmit && npx eslint src/components/ui/toast-provider.tsx
```
Expected: clean (no output).

---

### Task 4: Build the useToast hook and wire the provider into the root layout

**Files:**
- Create: `frontend/src/hooks/use-toast.ts`
- Modify: `frontend/src/app/layout.tsx`

**Interfaces:**
- Consumes: `ToastContext` from Task 3 (`@/components/ui/toast-provider`).
- Produces: `export function useToast(): ToastContextValue` — throws if called outside `<ToastProvider>`. Tasks 5 and 6 call this from Client Components.

- [ ] **Step 1: Write the hook**

```ts
"use client";

import { useContext } from "react";

import { ToastContext, type ToastContextValue } from "@/components/ui/toast-provider";

export function useToast(): ToastContextValue {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within a ToastProvider");
  }
  return context;
}
```

- [ ] **Step 2: Wrap the app in `frontend/src/app/layout.tsx`**

Add the import:
```ts
import ToastProvider from "@/components/ui/toast-provider";
```
Replace:
```tsx
      <body className="min-h-full flex flex-col">{children}</body>
```
with:
```tsx
      <body className="min-h-full flex flex-col">
        <ToastProvider>{children}</ToastProvider>
      </body>
```

- [ ] **Step 3: Verify types/lint**

```bash
cd frontend && npx tsc --noEmit && npx eslint src/hooks/use-toast.ts src/app/layout.tsx
```
Expected: clean (no output).

---

### Task 5: Wire LoginForm to use error toasts instead of the inline error paragraph

**Files:**
- Modify: `frontend/src/app/login/_lib/actions.ts`
- Modify: `frontend/src/app/login/_components/LoginForm.tsx`

**Interfaces:**
- Consumes: `useToast()` from Task 4.
- Produces: `LoginFormState` becomes `{ title?: string; message?: string; username?: string }` (was `{ error?: string; username?: string }`) — no other file reads `LoginFormState.error` today (verified: only `LoginForm.tsx` reads it).

- [ ] **Step 1: Update `frontend/src/app/login/_lib/actions.ts`**

Replace the whole file with:
```ts
"use server";

import { redirect } from "next/navigation";

import { login } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";

export interface LoginFormState {
  title?: string;
  message?: string;
  username?: string;
}

export async function loginAction(
  _prevState: LoginFormState,
  formData: FormData,
): Promise<LoginFormState> {
  const username = String(formData.get("username") ?? "").trim();
  const password = String(formData.get("password") ?? "");

  if (!username || !password) {
    return {
      title: "Invalid Format",
      message: "Enter your username and password.",
      username,
    };
  }

  try {
    await login({ username, password });
  } catch (err) {
    if (err instanceof ApiError) {
      return { title: err.title, message: err.message, username };
    }
    return {
      title: "Something Went Wrong",
      message: "Please try again.",
      username,
    };
  }

  redirect("/dashboard");
}
```

- [ ] **Step 2: Update `frontend/src/app/login/_components/LoginForm.tsx`**

Replace the whole file with:
```tsx
"use client";

import { useActionState, useEffect, useRef } from "react";

import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";

import { loginAction, type LoginFormState } from "../_lib/actions";

const initialState: LoginFormState = {};

export default function LoginForm() {
  const [state, formAction, isPending] = useActionState(
    loginAction,
    initialState,
  );
  const toast = useToast();
  const lastShown = useRef<LoginFormState | null>(null);

  useEffect(() => {
    if (state.title && state !== lastShown.current) {
      lastShown.current = state;
      toast.error(state.title, state.message ?? "");
    }
  }, [state, toast]);

  return (
    <form action={formAction} className="flex flex-col gap-5">
      <div>
        <Label htmlFor="username">Username</Label>
        <Input
          id="username"
          name="username"
          autoComplete="username"
          defaultValue={state.username}
          autoFocus
          required
        />
      </div>

      <div>
        <Label htmlFor="password">Password</Label>
        <Input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          required
        />
      </div>

      <Button type="submit" disabled={isPending} className="mt-1 w-full">
        {isPending ? "Signing in…" : "Sign in"}
      </Button>
    </form>
  );
}
```

Note: the `lastShown` ref guard prevents the effect from re-firing the same toast if `LoginForm` re-renders for an unrelated reason without a new action dispatch (each `loginAction` call produces a new `state` object identity, so this correctly fires once per submission).

- [ ] **Step 3: Verify types/lint**

```bash
cd frontend && npx tsc --noEmit && npx eslint src/app/login/_lib/actions.ts src/app/login/_components/LoginForm.tsx
```
Expected: clean (no output).

---

### Task 6: Wire a real success toast (logout → login page)

**Files:**
- Modify: `frontend/src/app/dashboard/_lib/actions.ts`
- Create: `frontend/src/app/login/_components/SignedOutNotice.tsx`
- Modify: `frontend/src/app/login/page.tsx`

**Interfaces:**
- Consumes: `useToast()` from Task 4.
- Produces: nothing consumed by later tasks (this is the last wiring task).

- [ ] **Step 1: Update `frontend/src/app/dashboard/_lib/actions.ts`**

Replace:
```ts
export async function logoutAction(): Promise<void> {
  await logout();
  redirect("/login");
}
```
with:
```ts
export async function logoutAction(): Promise<void> {
  await logout();
  redirect("/login?signedOut=1");
}
```

- [ ] **Step 2: Create `frontend/src/app/login/_components/SignedOutNotice.tsx`**

```tsx
"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useRef } from "react";

import { useToast } from "@/hooks/use-toast";

// Renders nothing - fires a one-shot success toast when arriving from
// logout, then strips the query param so a page refresh doesn't re-fire it.
export default function SignedOutNotice() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const toast = useToast();
  const fired = useRef(false);

  useEffect(() => {
    if (fired.current || searchParams.get("signedOut") !== "1") {
      return;
    }
    fired.current = true;
    toast.success("Signed out", "You've been signed out successfully.");
    router.replace("/login");
  }, [searchParams, router, toast]);

  return null;
}
```

- [ ] **Step 3: Render it in `frontend/src/app/login/page.tsx`**

Add imports:
```ts
import { Suspense } from "react";

import SignedOutNotice from "./_components/SignedOutNotice";
```
Replace:
```tsx
          <div className="mt-8">
            <LoginForm />
          </div>
```
with:
```tsx
          <div className="mt-8">
            <LoginForm />
          </div>

          <Suspense fallback={null}>
            <SignedOutNotice />
          </Suspense>
```

(`useSearchParams` requires a `Suspense` boundary per Next.js's App Router rules — see `node_modules/next/dist/docs/01-app/03-api-reference/04-functions/use-search-params.md`.)

- [ ] **Step 4: Verify types/lint**

```bash
cd frontend && npx tsc --noEmit && npx eslint src/app/dashboard/_lib/actions.ts src/app/login/_components/SignedOutNotice.tsx src/app/login/page.tsx
```
Expected: clean (no output).

---

### Task 7: End-to-end verification

**Files:** none (verification only).

- [ ] **Step 1: Full project typecheck + lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```
Expected: clean (no output / no errors).

- [ ] **Step 2: Rebuild and run the dev server, drive the real flow**

Start (or reuse) the frontend dev server and backend per this project's established local setup (see `frontend/.env.local` for `API_BASE_URL`). Using a headless-browser script (this project's established verification method — see prior session work using `playwright` in the scratchpad directory), drive:
1. Go to `/login`, submit wrong credentials → error toast appears (slides down from top-right) showing the backend's real title (e.g. "Unauthorized") and message (e.g. "Incorrect username or password."), with a red X-circle icon on the left and a red progress bar filling over ~5s, then the toast slides up and disappears on its own.
2. Submit again with wrong credentials, then click the toast's close (X) button before 5s — confirm it dismisses immediately (slide-up) instead of waiting.
3. Log in with valid credentials → redirected to `/dashboard`, no stray toast.
4. Click "Log out" → redirected to `/login`, a green success toast ("Signed out") appears and auto-dismisses; refreshing `/login` afterward must NOT re-show it (query param stripped).
5. Repeat 1 and 4 with the OS/browser color scheme set to dark — confirm both toast variants remain legible (background, text, icon, and progress-bar colors) per `AGENTS.md`'s "check in both light and dark" rule.
6. Repeat at a mobile viewport width (e.g. 390px) — confirm the toast stack doesn't overflow the viewport width and stays comfortably tappable (close button), per `AGENTS.md`'s responsiveness rule.

Expected: all six checks pass with no console errors.

## Self-Review Notes

- **Spec coverage:** icon library (Task 1), title(big)+message(below)+left icon+close button (Task 2), green=success/red=error (Task 2), 5s progress bar synced to auto-dismiss (Task 2+3), slide-down-in/slide-up-out (Task 2), replace existing error print (Task 5), reusable success+error API (Task 3/4), frontend-design visual polish applied throughout Task 2 (brand-consistent radii/shadow/typography) — all covered.
- **Type consistency:** `ToastVariant`, `ToastContextValue`, `LoginFormState` field names checked consistent across all tasks that reference them.
- **No git repo / no test runner:** confirmed via `package.json` (no `test` script) and repo root context; plan uses `tsc`/`eslint`/manual browser verification instead of unit tests and commits, matching this project's actual toolchain.
