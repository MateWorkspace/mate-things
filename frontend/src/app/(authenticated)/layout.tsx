import type { ReactNode } from "react";

import AppShell from "@/components/layout/AppShell";
import { getRoleById } from "@/lib/api/roles";
import { requireSessionContext } from "@/lib/session";

interface AuthenticatedLayoutProps {
  children: ReactNode;
}

export default async function AuthenticatedLayout({
  children,
}: AuthenticatedLayoutProps) {
  const session = await requireSessionContext();
  const roleName = session.permissions.has("role:get")
    ? await getRoleById(session.user.role_id)
        .then((role) => role.name)
        .catch(() => undefined)
    : undefined;

  return (
    <AppShell
      user={session.user}
      permissions={[...session.permissions]}
      roleName={roleName}
    >
      {children}
    </AppShell>
  );
}
