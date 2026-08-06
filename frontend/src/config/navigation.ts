import { canAccessAny, type PermissionName } from "@/lib/permissions";

export interface NavigationItem {
  label: string;
  href: string;
  requiredPermissions: readonly PermissionName[];
}

export interface NavigationGroup {
  label: string;
  items: readonly NavigationItem[];
}

export const NAVIGATION_GROUPS: readonly NavigationGroup[] = [
  {
    label: "Overview",
    items: [
      {
        label: "Fleet Overview",
        href: "/dashboard",
        requiredPermissions: [],
      },
    ],
  },
  {
    label: "Fleet",
    items: [
      { label: "Nodes", href: "/nodes", requiredPermissions: ["node:get"] },
      {
        label: "Node Classes",
        href: "/node-classes",
        requiredPermissions: ["node_class:get"],
      },
      {
        label: "Firmware",
        href: "/firmware",
        requiredPermissions: ["firmware:get"],
      },
    ],
  },
  {
    label: "Operations",
    items: [
      {
        label: "Actions",
        href: "/actions",
        requiredPermissions: ["action:get"],
      },
      {
        label: "Action History",
        href: "/action-history",
        requiredPermissions: ["action_log:get"],
      },
      {
        label: "BLE Direct",
        href: "/ble-direct",
        requiredPermissions: [],
      },
    ],
  },
  {
    label: "Observability",
    items: [
      {
        label: "Telemetry",
        href: "/telemetry",
        requiredPermissions: ["telemetry_record:get"],
      },
      {
        label: "Node Logs",
        href: "/node-logs",
        requiredPermissions: ["node_log:get"],
      },
      {
        label: "Broadcast Sessions",
        href: "/broadcast-sessions",
        requiredPermissions: ["broadcast_session:get"],
      },
    ],
  },
  {
    label: "Administration",
    items: [
      {
        label: "Users",
        href: "/admin/users",
        requiredPermissions: ["user:get"],
      },
      {
        label: "Access Control",
        href: "/admin/access-control",
        requiredPermissions: [
          "role:get",
          "permission:get",
          "role_permission:get",
        ],
      },
      {
        label: "Payload Schemas",
        href: "/admin/payload-schemas",
        requiredPermissions: ["payload_schema:get"],
      },
    ],
  },
];

export function visibleNavigation(
  permissions: ReadonlySet<PermissionName>,
): NavigationGroup[] {
  return NAVIGATION_GROUPS.flatMap((group) => {
    const items = group.items.filter((item) =>
      canAccessAny(permissions, item.requiredPermissions),
    );

    return items.length === 0 ? [] : [{ ...group, items }];
  });
}
