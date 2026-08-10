import { canVisitRoute, findRoutePolicy } from "@/config/route-policies";
import type { PermissionName } from "@/lib/permissions";

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
  return findRoutePolicy(normalized)?.requiredAny ?? [];
}

export function canVisit(
  pathname: string,
  permissions: ReadonlySet<PermissionName>,
): boolean {
  return canVisitRoute(normalizePathname(pathname), permissions);
}
