import {
  canVisitRoute,
  ROUTE_POLICIES,
  type RoutePolicy,
} from "@/config/route-policies";
import type { PermissionName } from "@/lib/permissions";

export interface NavigationItem {
  label: string;
  href: string;
  requiredPermissions: readonly PermissionName[];
}

export interface NavigationGroup {
  label: string;
  items: readonly NavigationItem[];
}

const NAVIGATION_GROUP_PRESENTATION: readonly {
  group: NonNullable<RoutePolicy["navigation"]>["group"];
  label: string;
}[] = [
  { group: "overview", label: "Overview" },
  { group: "fleet", label: "Fleet" },
  { group: "operations", label: "Operations" },
  { group: "observability", label: "Observability" },
  { group: "administration", label: "Administration" },
];

export const NAVIGATION_GROUPS: readonly NavigationGroup[] =
  NAVIGATION_GROUP_PRESENTATION.flatMap(({ group, label }) => {
    const items = ROUTE_POLICIES.flatMap((policy) => {
      if (policy.navigation?.group !== group) {
        return [];
      }

      return [
        {
          label: policy.navigation.label,
          href: policy.href,
          requiredPermissions: policy.requiredAny ?? [],
        },
      ];
    });

    return items.length === 0 ? [] : [{ label, items }];
  });

export function visibleNavigation(
  permissions: ReadonlySet<PermissionName>,
): NavigationGroup[] {
  return NAVIGATION_GROUPS.flatMap((group) => {
    const items = group.items.filter((item) =>
      canVisitRoute(item.href, permissions),
    );

    return items.length === 0 ? [] : [{ ...group, items }];
  });
}
