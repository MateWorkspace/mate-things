# Route map

The app uses the Next.js 16 App Router.

| URL | Entry | Layout | Current purpose |
|---|---|---|---|
| `/` | `src/app/page.tsx` | `src/app/layout.tsx` | Redirects authenticated users to `/dashboard`, otherwise `/login`. |
| `/login` | `src/app/login/page.tsx` | `src/app/layout.tsx` | Branded username/password sign-in page. |
| `/dashboard` | `src/app/dashboard/page.tsx` | `src/app/layout.tsx` | Protected placeholder dashboard showing the signed-in user and logout. |

`src/proxy.ts` refreshes JWT cookies and currently protects only `/dashboard`.

## Backend-aligned target areas

- Overview: fleet health dashboard.
- Fleet: nodes, node details/config/OTA, node classes, firmware.
- Operations: actions, action dispatch, action history.
- Observability: telemetry and node logs.
- Administration: users, roles, permissions, payload schemas.
- Account: profile details, editing, and password security.
