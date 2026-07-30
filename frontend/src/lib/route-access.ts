import type { PermissionName } from "@/lib/permissions";

interface RouteAccessRule {
  path: string;
  permissions: readonly PermissionName[];
}

const ROUTE_ACCESS_RULES: readonly RouteAccessRule[] = [
  {
    path: "/admin/access-control",
    permissions: ["role:get", "permission:get", "role_permission:get"],
  },
  { path: "/admin/payload-schemas", permissions: ["payload_schema:get"] },
  { path: "/admin/users", permissions: ["user:get"] },
  { path: "/action-history", permissions: ["action_log:get"] },
  { path: "/node-classes", permissions: ["node_class:get"] },
  { path: "/node-logs", permissions: ["node_log:get"] },
  { path: "/telemetry", permissions: ["telemetry_record:get"] },
  { path: "/firmware", permissions: ["firmware:get"] },
  { path: "/actions", permissions: ["action:get"] },
  { path: "/nodes", permissions: ["node:get"] },
  { path: "/dashboard", permissions: [] },
];

function normalizePathname(pathname: string): string {
  const path = pathname.split(/[?#]/, 1)[0] || "/";
  if (path === "/") {
    return path;
  }
  return path.replace(/\/+$/, "");
}

export function requiredPermissions(
  pathname: string,
): readonly PermissionName[] {
  const normalized = normalizePathname(pathname);
  return (
    ROUTE_ACCESS_RULES.find(
      (rule) =>
        normalized === rule.path || normalized.startsWith(`${rule.path}/`),
    )?.permissions ?? []
  );
}

export function canVisit(
  pathname: string,
  permissions: ReadonlySet<PermissionName>,
): boolean {
  const required = requiredPermissions(pathname);
  return (
    required.length === 0 ||
    required.some((permission) => permissions.has(permission))
  );
}
