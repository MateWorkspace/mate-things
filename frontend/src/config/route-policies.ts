import type { PermissionName } from "@/lib/permissions";

export type RoutePolicy = {
  href: string;
  protected: boolean;
  requiredAny?: readonly PermissionName[];
  navigation?: {
    group:
      | "overview"
      | "fleet"
      | "operations"
      | "observability"
      | "applications"
      | "administration";
    app?: { key: string; label: string };
    label: string;
  };
};

export const ROUTE_POLICIES: readonly RoutePolicy[] = [
  {
    href: "/",
    protected: false,
  },
  {
    href: "/login",
    protected: false,
  },
  {
    href: "/auth/invalid-session",
    protected: false,
  },
  {
    href: "/dashboard",
    protected: true,
    requiredAny: [],
    navigation: { group: "overview", label: "Fleet Overview" },
  },
  {
    href: "/nodes",
    protected: true,
    requiredAny: ["node:get"],
    navigation: { group: "fleet", label: "Nodes" },
  },
  {
    href: "/node-classes",
    protected: true,
    requiredAny: ["node_class:get"],
    navigation: { group: "fleet", label: "Node Classes" },
  },
  {
    href: "/firmware",
    protected: true,
    requiredAny: ["firmware:get"],
    navigation: { group: "fleet", label: "Firmware" },
  },
  {
    href: "/actions",
    protected: true,
    requiredAny: ["action:get"],
    navigation: { group: "operations", label: "Actions" },
  },
  {
    href: "/action-history",
    protected: true,
    requiredAny: ["action_log:get"],
    navigation: { group: "operations", label: "Action History" },
  },
  {
    href: "/ble-direct",
    protected: true,
    requiredAny: [],
    navigation: { group: "operations", label: "BLE Direct" },
  },
  {
    href: "/telemetry",
    protected: true,
    requiredAny: ["telemetry_record:get"],
    navigation: { group: "observability", label: "Telemetry" },
  },
  {
    href: "/node-logs",
    protected: true,
    requiredAny: ["node_log:get"],
    navigation: { group: "observability", label: "Node Logs" },
  },
  {
    href: "/broadcast-sessions",
    protected: true,
    requiredAny: ["broadcast_session:get"],
    navigation: {
      group: "observability",
      label: "Broadcast Sessions",
    },
  },
  { href: "/apps/infrared", protected: true },
  {
    href: "/apps/infrared/settings",
    protected: true,
    requiredAny: ["infrared_reference:get"],
    navigation: {
      group: "applications",
      app: { key: "infrared", label: "Infrared" },
      label: "Settings",
    },
  },
  {
    href: "/apps/infrared/record",
    protected: true,
    requiredAny: ["infrared_record_session:get"],
    navigation: {
      group: "applications",
      app: { key: "infrared", label: "Infrared" },
      label: "Record",
    },
  },
  {
    href: "/apps/infrared/command",
    protected: true,
    requiredAny: ["infrared_record_session:get"],
    navigation: {
      group: "applications",
      app: { key: "infrared", label: "Infrared" },
      label: "Command",
    },
  },
  { href: "/admin", protected: true },
  {
    href: "/admin/users",
    protected: true,
    requiredAny: ["user:get"],
    navigation: { group: "administration", label: "Users" },
  },
  {
    href: "/admin/api-keys",
    protected: true,
    requiredAny: ["api_key:get"],
    navigation: { group: "administration", label: "API Keys" },
  },
  {
    href: "/admin/access-control",
    protected: true,
    requiredAny: ["role:get", "permission:get", "role_permission:get"],
    navigation: { group: "administration", label: "Access Control" },
  },
  {
    href: "/admin/payload-schemas",
    protected: true,
    requiredAny: ["payload_schema:get"],
    navigation: { group: "administration", label: "Payload Schemas" },
  },
  {
    href: "/admin/llm-config",
    protected: true,
    requiredAny: ["llm_config:get"],
    navigation: { group: "administration", label: "LLM Config" },
  },
];

const matchesPrefix = (pathname: string, href: string) =>
  pathname === href || pathname.startsWith(`${href}/`);

export function findRoutePolicy(pathname: string): RoutePolicy | undefined {
  return [...ROUTE_POLICIES]
    .sort((left, right) => right.href.length - left.href.length)
    .find((policy) => matchesPrefix(pathname, policy.href));
}

export function isProtectedRoute(pathname: string): boolean {
  const policy = findRoutePolicy(pathname);
  return policy ? policy.protected : true;
}

export function canVisitRoute(
  pathname: string,
  permissions: ReadonlySet<string>,
): boolean {
  const policy = findRoutePolicy(pathname);
  if (!policy) {
    return false;
  }

  const requiredAny = policy.requiredAny ?? [];

  return (
    requiredAny.length === 0 ||
    requiredAny.some((permission) => permissions.has(permission))
  );
}
