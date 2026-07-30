import Image from "next/image";
import { Menu, PanelLeftClose, PanelLeftOpen, UserRound } from "lucide-react";

import hat from "@/assets/hat.svg";
import IconButton from "@/components/ui/icon-button";
import type { UserResponse } from "@/lib/api/users";

interface AppBarProps {
  user: UserResponse;
  sidebarCollapsed: boolean;
  onDesktopSidebarToggle: () => void;
  onMobileNavigationOpen: () => void;
}

export default function AppBar({
  user,
  sidebarCollapsed,
  onDesktopSidebarToggle,
  onMobileNavigationOpen,
}: AppBarProps) {
  return (
    <header className="border-border bg-background/95 sticky top-0 z-30 flex h-16 items-center justify-between gap-4 border-b px-4 backdrop-blur sm:px-6">
      <div className="flex min-w-0 items-center gap-2">
        <IconButton
          aria-label="Open navigation"
          className="lg:hidden"
          onClick={onMobileNavigationOpen}
        >
          <Menu aria-hidden="true" className="size-5" />
        </IconButton>
        <IconButton
          aria-label={sidebarCollapsed ? "Expand sidebar" : "Collapse sidebar"}
          className="hidden lg:inline-flex"
          onClick={onDesktopSidebarToggle}
        >
          {sidebarCollapsed ? (
            <PanelLeftOpen aria-hidden="true" className="size-5" />
          ) : (
            <PanelLeftClose aria-hidden="true" className="size-5" />
          )}
        </IconButton>
        <div className="flex min-w-0 items-center gap-2">
          <Image
            src={hat}
            alt=""
            width={32}
            height={32}
            className="size-8 shrink-0"
            priority
          />
          <span className="font-display text-primary hidden truncate text-lg tracking-wide sm:inline">
            Mate Things
          </span>
        </div>
      </div>

      <button
        type="button"
        className="hover:bg-muted focus-visible:ring-focus focus-visible:ring-offset-background flex min-h-11 min-w-0 items-center gap-2 rounded-xl px-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
        aria-label={`Open profile for ${user.name}`}
      >
        <span className="hidden max-w-40 truncate sm:inline">{user.name}</span>
        <span className="bg-surface text-primary flex size-9 shrink-0 items-center justify-center rounded-full">
          <UserRound aria-hidden="true" className="size-5" />
        </span>
      </button>
    </header>
  );
}
