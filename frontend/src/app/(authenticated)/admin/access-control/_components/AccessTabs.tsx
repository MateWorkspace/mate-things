import Link from "next/link";

export default function AccessTabs({
  active,
  showRoles,
  showPermissions,
}: {
  active: "roles" | "permissions";
  showRoles: boolean;
  showPermissions: boolean;
}) {
  return (
    <nav
      aria-label="Access control resources"
      className="border-border flex gap-2 border-b"
    >
      {showRoles ? (
        <Link
          aria-current={active === "roles" ? "page" : undefined}
          href="/admin/access-control?tab=roles"
          className={`rounded-t-xl border-b-2 px-4 py-3 text-sm font-semibold ${active === "roles" ? "border-primary text-primary" : "text-muted-foreground border-transparent"}`}
        >
          Roles
        </Link>
      ) : null}
      {showPermissions ? (
        <Link
          aria-current={active === "permissions" ? "page" : undefined}
          href="/admin/access-control?tab=permissions"
          className={`rounded-t-xl border-b-2 px-4 py-3 text-sm font-semibold ${active === "permissions" ? "border-primary text-primary" : "text-muted-foreground border-transparent"}`}
        >
          Permissions
        </Link>
      ) : null}
    </nav>
  );
}
