import type { PermissionResponse } from "@/lib/api/permissions";

const GROUPS = [
  "Profile",
  "Fleet",
  "Operations",
  "Observability",
  "Administration",
] as const;

function groupFor(name: string): (typeof GROUPS)[number] {
  const resource = name.split(":")[0];
  if (resource === "profile") return "Profile";
  if (
    ["node", "node_class", "node_config", "firmware", "ota"].includes(resource)
  )
    return "Fleet";
  if (["action", "action_log"].includes(resource)) return "Operations";
  if (["telemetry_record", "node_log"].includes(resource))
    return "Observability";
  return "Administration";
}

export default function PermissionGroups({
  permissions,
  selected,
  editable,
  onSelectionChange,
}: {
  permissions: readonly PermissionResponse[];
  selected: ReadonlySet<string>;
  editable: boolean;
  onSelectionChange: (id: string, selected: boolean) => void;
}) {
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      {GROUPS.map((group) => {
        const items = permissions.filter(
          (permission) => groupFor(permission.name) === group,
        );
        if (!items.length) return null;
        return (
          <fieldset
            key={group}
            aria-label={group}
            className="border-border rounded-xl border p-4"
          >
            <legend className="font-display text-primary px-2 text-lg">
              {group}
            </legend>
            <div className="space-y-2">
              {items.map((permission) => (
                <label
                  key={permission.id}
                  className="bg-muted flex gap-3 rounded-lg p-3"
                >
                  {editable ? (
                    <input
                      type="checkbox"
                      name="permission_ids"
                      value={permission.id}
                      checked={selected.has(permission.id)}
                      onChange={(event) =>
                        onSelectionChange(permission.id, event.target.checked)
                      }
                      className="accent-primary mt-1 size-4"
                    />
                  ) : (
                    <span aria-hidden="true" className="mt-1">
                      {selected.has(permission.id) ? "✓" : "—"}
                    </span>
                  )}
                  <span>
                    <span className="block font-mono text-sm font-semibold">
                      {permission.name}
                    </span>
                    <span className="text-muted-foreground text-xs">
                      {permission.description || "No description."}
                    </span>
                  </span>
                </label>
              ))}
            </div>
          </fieldset>
        );
      })}
    </div>
  );
}
