"use client";

import { useRouter } from "next/navigation";
import { useActionState, useEffect, useState } from "react";

import PreferencesDialog from "@/components/preferences/PreferencesDialog";
import ResourceCard from "@/components/collection/ResourceCard";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { PermissionResponse } from "@/lib/api/permissions";

import { removePermissionAction, savePermissionAction } from "../_lib/actions";
import { EMPTY_ACCESS_STATE } from "../_lib/state";

export default function PermissionCard({
  permission,
  grants,
}: {
  permission: PermissionResponse;
  grants: readonly string[];
}) {
  const allowed = new Set(grants);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  return (
    <ResourceCard title={permission.name}>
      <p className="text-foreground/70 min-h-15">
        {permission.description || "No description."}
      </p>
      <div className="mt-4 flex flex-wrap gap-2">
        {allowed.has("permission:set") ? (
          <Button variant="secondary" onClick={() => setEditOpen(true)}>
            Edit
          </Button>
        ) : null}
        <PreferencesDialog
          resource="permission"
          id={permission.id}
          preferences={permission.preferences}
          permissions={grants}
          label="Preferences"
        />
        {allowed.has("permission:remove") ? (
          <Button variant="critical" onClick={() => setDeleteOpen(true)}>
            Delete
          </Button>
        ) : null}
      </div>
      <PermissionDialog
        permission={permission}
        open={editOpen}
        onClose={() => setEditOpen(false)}
        title={`Edit ${permission.name}`}
      />
      <DeletePermissionDialog
        permission={permission}
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
      />
    </ResourceCard>
  );
}

export function PermissionDialog({
  permission,
  open,
  onClose,
  title,
}: {
  permission?: PermissionResponse;
  open: boolean;
  onClose: () => void;
  title: string;
}) {
  const [state, action, pending] = useActionState(
    savePermissionAction,
    EMPTY_ACCESS_STATE,
  );
  const router = useRouter();
  useEffect(() => {
    if (state.status === "success") {
      router.refresh();
    }
  }, [state, router]);
  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={title}
      variant="sheet"
    >
      <form action={action} className="space-y-4">
        {permission ? (
          <input type="hidden" name="permission_id" value={permission.id} />
        ) : null}
        <div>
          <Label htmlFor="permission-name">Permission name</Label>
          <Input
            id="permission-name"
            name="name"
            defaultValue={permission?.name}
            placeholder="resource:operation"
            required
          />
        </div>
        <div>
          <Label htmlFor="permission-description">Description</Label>
          <textarea
            id="permission-description"
            name="description"
            defaultValue={permission?.description}
            rows={4}
            className="border-control-border bg-background w-full rounded-xl border p-3 text-sm"
          />
        </div>
        <p
          className={
            state.status === "error"
              ? "text-critical text-sm"
              : "text-success text-sm"
          }
        >
          {state.message}
        </p>
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Saving…" : "Save permission"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

function DeletePermissionDialog({
  permission,
  open,
  onClose,
}: {
  permission: PermissionResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [state, action, pending] = useActionState(
    removePermissionAction,
    EMPTY_ACCESS_STATE,
  );
  const router = useRouter();
  useEffect(() => {
    if (state.status === "success") {
      router.refresh();
    }
  }, [state, router]);
  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={`Delete ${permission.name}`}
      variant="sheet"
    >
      <form action={action} className="space-y-4">
        <input type="hidden" name="permission_id" value={permission.id} />
        <input type="hidden" name="permission_name" value={permission.name} />
        <p className="bg-muted rounded-xl p-4 text-sm">
          Existing role assignments can cause the backend to reject this
          deletion. Enter <strong>{permission.name}</strong> to continue.
        </p>
        <Input
          name="confirmation"
          aria-label="Confirm permission name"
          required
        />
        <p
          className={
            state.status === "error"
              ? "text-critical text-sm"
              : "text-success text-sm"
          }
        >
          {state.message}
        </p>
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Deleting…" : "Delete permission"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
