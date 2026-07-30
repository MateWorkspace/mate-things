import { cookies } from "next/headers";
import type { ReactNode } from "react";

import AppShell from "@/components/layout/AppShell";
import { requireSessionContext } from "@/lib/session";

const SIDEBAR_COLLAPSED_COOKIE = "mate_sidebar_collapsed";

interface AuthenticatedLayoutProps {
  children: ReactNode;
}

export default async function AuthenticatedLayout({
  children,
}: AuthenticatedLayoutProps) {
  const [session, cookieStore] = await Promise.all([
    requireSessionContext(),
    cookies(),
  ]);

  return (
    <AppShell
      user={session.user}
      permissions={[...session.permissions]}
      initialSidebarCollapsed={
        cookieStore.get(SIDEBAR_COLLAPSED_COOKIE)?.value === "1"
      }
    >
      {children}
    </AppShell>
  );
}
