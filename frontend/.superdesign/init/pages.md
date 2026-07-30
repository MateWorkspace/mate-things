# Page dependency trees

## `/` (root redirect)

Entry: `src/app/page.tsx`

- `src/lib/session/index.ts`
  - `src/lib/session/jwt.ts`
  - `src/lib/session/session.ts`
    - `src/lib/api/client.ts`
    - `src/lib/api/profile.ts`

## `/login`

Entry: `src/app/login/page.tsx`

- `src/assets/matethings-horizontal.svg`
- `src/app/login/_components/LoginForm.tsx`
  - `src/components/ui/button.tsx`
  - `src/components/ui/input.tsx`
  - `src/components/ui/label.tsx`
  - `src/hooks/use-toast.ts`
    - `src/components/ui/toast-provider.tsx`
      - `src/components/ui/toast.tsx`
  - `src/app/login/_lib/actions.ts`
    - `src/lib/api/auth.ts`
    - `src/lib/api/client.ts`
- `src/app/login/_components/SignedOutNotice.tsx`
  - `src/hooks/use-toast.ts`

## `/dashboard`

Entry: `src/app/dashboard/page.tsx`

- `src/assets/hat.svg`
- `src/lib/session/index.ts`
  - `src/lib/session/session.ts`
    - `src/lib/api/profile.ts`
    - `src/lib/api/client.ts`
- `src/app/dashboard/_components/LogoutButton.tsx`
  - `src/components/ui/button.tsx`
  - `src/app/dashboard/_lib/actions.ts`
    - `src/lib/api/auth.ts`

## Global root layout

Entry: `src/app/layout.tsx`

- `src/app/globals.css`
- `src/components/ui/toast-provider.tsx`
  - `src/components/ui/toast.tsx`
