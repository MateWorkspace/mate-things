"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import { usePathname } from "next/navigation";

import { visibleNavigation } from "@/config/navigation";
import Dialog from "@/components/ui/dialog";
import { setSidebarCollapsedAction } from "@/lib/api/session-actions";
import type { UserResponse } from "@/lib/api/users";

import AppBar from "./AppBar";
import Sidebar from "./Sidebar";

interface AppShellProps {
  user: UserResponse;
  permissions: readonly string[];
  children: ReactNode;
  initialSidebarCollapsed?: boolean;
}

export default function AppShell({
  user,
  permissions,
  children,
  initialSidebarCollapsed = false,
}: AppShellProps) {
  const pathname = usePathname();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(
    initialSidebarCollapsed,
  );
  const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false);
  const sidebarPreferenceWrite = useRef(Promise.resolve());
  const navigation = visibleNavigation(new Set(permissions));

  useEffect(() => {
    setMobileNavigationOpen(false);
  }, [pathname]);

  const toggleDesktopSidebar = () => {
    const nextCollapsed = !sidebarCollapsed;
    setSidebarCollapsed(nextCollapsed);
    sidebarPreferenceWrite.current = sidebarPreferenceWrite.current
      .catch(() => undefined)
      .then(() => setSidebarCollapsedAction(nextCollapsed));
  };

  return (
    <div className="bg-background min-h-screen overflow-x-clip">
      <a
        href="#main-content"
        className="bg-primary text-surface focus-visible:ring-focus fixed top-2 left-4 z-50 -translate-y-20 rounded-xl px-4 py-2.5 text-sm font-semibold transition-transform focus-visible:translate-y-0 focus-visible:ring-2 focus-visible:outline-none"
      >
        Skip to main content
      </a>
      <AppBar
        user={user}
        permissions={permissions}
        sidebarCollapsed={sidebarCollapsed}
        onDesktopSidebarToggle={toggleDesktopSidebar}
        onMobileNavigationOpen={() => setMobileNavigationOpen(true)}
      />
      <div className="flex min-h-[calc(100vh-4rem)]">
        <aside
          className={`border-border bg-surface/30 hidden shrink-0 border-r transition-[width] lg:block ${
            sidebarCollapsed ? "w-20" : "w-64"
          }`}
        >
          <Sidebar
            navigation={navigation}
            pathname={pathname}
            collapsed={sidebarCollapsed}
            idPrefix="desktop"
          />
        </aside>
        <div
          id="main-content"
          tabIndex={-1}
          className="min-w-0 flex-1 focus:outline-none"
        >
          {children}
        </div>
      </div>
      <Dialog
        open={mobileNavigationOpen}
        onClose={() => setMobileNavigationOpen(false)}
        title="Navigation"
        variant="drawer"
      >
        <Sidebar
          navigation={navigation}
          pathname={pathname}
          collapsed={false}
          idPrefix="mobile"
          onNavigate={() => setMobileNavigationOpen(false)}
        />
      </Dialog>
    </div>
  );
}
