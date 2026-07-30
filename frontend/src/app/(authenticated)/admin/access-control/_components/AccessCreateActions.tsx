"use client";

import { useState } from "react";

import Button from "@/components/ui/button";

import { PermissionDialog } from "./PermissionCard";
import { RoleDialog } from "./RoleDetails";
import { saveRoleAction } from "../_lib/actions";

export default function AccessCreateActions({
  canAddRole,
  canAddPermission,
  tab,
}: {
  canAddRole: boolean;
  canAddPermission: boolean;
  tab: "roles" | "permissions";
}) {
  const [open, setOpen] = useState(false);
  if (
    (tab === "roles" && !canAddRole) ||
    (tab === "permissions" && !canAddPermission)
  )
    return null;
  return (
    <>
      <Button onClick={() => setOpen(true)}>
        Create {tab === "roles" ? "role" : "permission"}
      </Button>
      {tab === "roles" ? (
        <RoleDialog
          open={open}
          onClose={() => setOpen(false)}
          action={saveRoleAction}
          title="Create role"
          mode="create"
        />
      ) : (
        <PermissionDialog
          open={open}
          onClose={() => setOpen(false)}
          title="Create permission"
        />
      )}
    </>
  );
}
