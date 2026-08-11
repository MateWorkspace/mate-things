"use client";

import Link from "next/link";
import { useState } from "react";
import { ChevronRight, Radio, Rss, type LucideIcon } from "lucide-react";

import type { NavApp } from "@/config/navigation";

import { isCurrentPath } from "./is-current-path";

const APP_ICONS: Readonly<Record<string, LucideIcon>> = {
  infrared: Rss,
};

interface SidebarAppProps {
  app: NavApp;
  pathname: string;
  idPrefix: string;
  onNavigate?: () => void;
}

export default function SidebarApp({
  app,
  pathname,
  idPrefix,
  onNavigate,
}: SidebarAppProps) {
  const hasActiveChild = app.pages.some((item) =>
    isCurrentPath(pathname, item.href),
  );
  const [expanded, setExpanded] = useState(hasActiveChild);
  const Icon = APP_ICONS[app.key] ?? Radio;
  const panelId = `${idPrefix}-app-${app.key}`;

  return (
    <div>
      <button
        type="button"
        aria-expanded={expanded}
        aria-controls={panelId}
        onClick={() => setExpanded((current) => !current)}
        className={`focus-visible:ring-focus flex min-h-11 w-full items-center gap-3 rounded-xl px-3 text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:outline-none ${
          hasActiveChild
            ? "text-primary"
            : "text-foreground/75 hover:bg-muted hover:text-foreground"
        }`}
      >
        <Icon aria-hidden="true" className="size-5 shrink-0" />
        <span className="flex-1 text-left">{app.label}</span>
        <ChevronRight
          aria-hidden="true"
          className={`size-4 shrink-0 transition-transform ${expanded ? "rotate-90" : ""}`}
        />
      </button>
      {expanded ? (
        <ul
          id={panelId}
          className="border-border mt-1 ml-[22px] space-y-1 border-l pl-4"
        >
          {app.pages.map((item) => {
            const current = isCurrentPath(pathname, item.href);

            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  aria-current={current ? "page" : undefined}
                  onClick={onNavigate}
                  className={`focus-visible:ring-focus flex min-h-9 items-center rounded-xl px-3 text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:outline-none ${
                    current
                      ? "bg-highlight/55 text-primary"
                      : "text-foreground/75 hover:bg-muted hover:text-foreground"
                  }`}
                >
                  {item.label}
                </Link>
              </li>
            );
          })}
        </ul>
      ) : null}
    </div>
  );
}
