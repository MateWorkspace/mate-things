import type { PermissionName } from "@/lib/permissions";

export interface NavPage {
  href: string;
  label: string;
  /** Omitted = any logged-in user may visit; otherwise any one permission suffices. */
  requiredAny?: readonly PermissionName[];
}

export interface NavApp {
  key: string;
  label: string;
  pages: readonly NavPage[];
}

export interface NavGroup {
  label: string;
  pages?: readonly NavPage[];
  apps?: readonly NavApp[];
}

/**
 * Exact-match (prefix-match, see lib/navigation.ts) routes that don't
 * require a session. Everything else defaults to protected, fail-closed.
 */
export const PUBLIC_ROUTES: readonly string[] = [
  "/",
  "/login",
  "/auth/invalid-session",
];

export const NAVIGATION: readonly NavGroup[] = [
  {
    label: "Overview",
    pages: [{ href: "/dashboard", label: "Fleet Overview" }],
  },
  {
    label: "Fleet",
    pages: [
      { href: "/nodes", label: "Nodes", requiredAny: ["node:get"] },
      {
        href: "/node-classes",
        label: "Node Classes",
        requiredAny: ["node_class:get"],
      },
      { href: "/firmware", label: "Firmware", requiredAny: ["firmware:get"] },
    ],
  },
  {
    label: "Operations",
    pages: [
      { href: "/actions", label: "Actions", requiredAny: ["action:get"] },
      {
        href: "/action-history",
        label: "Action History",
        requiredAny: ["action_log:get"],
      },
      { href: "/ble-direct", label: "BLE Direct" },
    ],
  },
  {
    label: "Observability",
    pages: [
      {
        href: "/telemetry",
        label: "Telemetry",
        requiredAny: ["telemetry_record:get"],
      },
      {
        href: "/node-logs",
        label: "Node Logs",
        requiredAny: ["node_log:get"],
      },
      {
        href: "/broadcast-sessions",
        label: "Broadcast Sessions",
        requiredAny: ["broadcast_session:get"],
      },
    ],
  },
  {
    label: "Applications",
    apps: [
      {
        key: "infrared",
        label: "Infrared",
        pages: [
          {
            href: "/apps/infrared/settings",
            label: "Settings",
            requiredAny: ["infrared_reference:get"],
          },
          {
            href: "/apps/infrared/record",
            label: "Record",
            requiredAny: ["infrared_record_session:get"],
          },
          {
            href: "/apps/infrared/command",
            label: "Command",
            requiredAny: ["infrared_record_session:get"],
          },
        ],
      },
    ],
  },
  {
    label: "Administration",
    pages: [
      { href: "/admin/users", label: "Users", requiredAny: ["user:get"] },
      {
        href: "/admin/api-keys",
        label: "API Keys",
        requiredAny: ["api_key:get"],
      },
      {
        href: "/admin/access-control",
        label: "Access Control",
        requiredAny: ["role:get", "permission:get", "role_permission:get"],
      },
      {
        href: "/admin/payload-schemas",
        label: "Payload Schemas",
        requiredAny: ["payload_schema:get"],
      },
      {
        href: "/admin/llm-config",
        label: "LLM Config",
        requiredAny: ["llm_config:get"],
      },
    ],
  },
];
