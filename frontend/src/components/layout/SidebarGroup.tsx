import Link from "next/link";
import {
  Activity,
  Boxes,
  Cpu,
  FileJson,
  History,
  LayoutDashboard,
  Radio,
  ScrollText,
  Server,
  Shield,
  Users,
  Workflow,
  type LucideIcon,
} from "lucide-react";

import type { NavigationGroup } from "@/config/navigation";

const NAVIGATION_ICONS: Readonly<Record<string, LucideIcon>> = {
  "/dashboard": LayoutDashboard,
  "/nodes": Server,
  "/node-classes": Boxes,
  "/firmware": Cpu,
  "/actions": Workflow,
  "/action-history": History,
  "/telemetry": Activity,
  "/node-logs": ScrollText,
  "/admin/users": Users,
  "/admin/access-control": Shield,
  "/admin/payload-schemas": FileJson,
};

interface SidebarGroupProps {
  group: NavigationGroup;
  pathname: string;
  collapsed: boolean;
  idPrefix: string;
  onNavigate?: () => void;
}

function isCurrentPath(pathname: string, href: string): boolean {
  return pathname === href || pathname.startsWith(`${href}/`);
}

export default function SidebarGroup({
  group,
  pathname,
  collapsed,
  idPrefix,
  onNavigate,
}: SidebarGroupProps) {
  const headingId = `${idPrefix}-navigation-${group.label
    .toLowerCase()
    .replaceAll(" ", "-")}`;

  return (
    <section aria-labelledby={headingId}>
      <h2
        id={headingId}
        className={`text-foreground/55 px-3 text-xs font-semibold tracking-wider uppercase ${
          collapsed ? "sr-only" : ""
        }`}
      >
        {group.label}
      </h2>
      <ul className="mt-2 space-y-1">
        {group.items.map((item) => {
          const Icon = NAVIGATION_ICONS[item.href] ?? Radio;
          const current = isCurrentPath(pathname, item.href);

          return (
            <li key={item.href}>
              <Link
                href={item.href}
                aria-current={current ? "page" : undefined}
                title={collapsed ? item.label : undefined}
                onClick={onNavigate}
                className={`focus-visible:ring-focus flex min-h-11 items-center rounded-xl px-3 text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:outline-none ${
                  current
                    ? "bg-highlight/55 text-primary"
                    : "text-foreground/75 hover:bg-muted hover:text-foreground"
                } ${collapsed ? "justify-center" : "gap-3"}`}
              >
                <Icon aria-hidden="true" className="size-5 shrink-0" />
                <span className={collapsed ? "sr-only" : ""}>{item.label}</span>
              </Link>
            </li>
          );
        })}
      </ul>
    </section>
  );
}
