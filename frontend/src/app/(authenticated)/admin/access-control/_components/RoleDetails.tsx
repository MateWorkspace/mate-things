"use client";

import { useActionState, useState } from "react";

import PreferencesDialog from "@/components/preferences/PreferencesDialog";
import Button from "@/components/ui/button";
import Card from "@/components/ui/card";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { PermissionResponse } from "@/lib/api/permissions";
import type { RoleResponse } from "@/lib/api/roles";

import {
  removeRoleAction,
  saveRoleAction,
  setDefaultRoleAction,
  updateRoleAssignmentsAction,
} from "../_lib/actions";
import { EMPTY_ACCESS_STATE } from "../_lib/state";
import PermissionGroups from "./PermissionGroups";

export default function RoleDetails({
  role,
  permissions,
  selected,
  grants,
}: {
  role: RoleResponse;
  permissions: readonly PermissionResponse[];
  selected: readonly string[];
  grants: readonly string[];
}) {
  const allowed = new Set(grants);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [defaultOpen, setDefaultOpen] = useState(false);
  const [assignmentState, assignmentAction, assignmentPending] = useActionState(
    updateRoleAssignmentsAction,
    EMPTY_ACCESS_STATE,
  );
  return (
    <section className="space-y-5">
      <Card>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="font-display text-primary text-2xl">{role.name}</h2>
            <p className="text-muted-foreground mt-1">
              {role.description || "No description."}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            {allowed.has("role:set") ? (
              <Button variant="secondary" onClick={() => setEditOpen(true)}>
                Edit role
              </Button>
            ) : null}
            {!role.is_default && allowed.has("role:set") ? (
              <Button variant="secondary" onClick={() => setDefaultOpen(true)}>
                Make default
              </Button>
            ) : null}
            {allowed.has("role:remove") ? (
              <button
                className="border-critical text-critical rounded-xl border px-4 text-sm font-semibold"
                onClick={() => setDeleteOpen(true)}
              >
                Delete role
              </button>
            ) : null}
            <PreferencesDialog
              resource="role"
              id={role.id}
              preferences={role.preferences}
              permissions={grants}
            />
          </div>
        </div>
      </Card>
      {permissions.length ? (
        <form action={assignmentAction} className="space-y-4">
          <input type="hidden" name="role_id" value={role.id} />
          <PermissionGroups
            permissions={permissions}
            selected={new Set(selected)}
            editable={
              allowed.has("role_permission:get") &&
              (allowed.has("role_permission:add") ||
                allowed.has("role_permission:remove"))
            }
          />
          <p
            aria-live="polite"
            className={
              assignmentState.status === "error"
                ? "text-critical text-sm"
                : "text-success text-sm"
            }
          >
            {assignmentState.message}
          </p>
          {allowed.has("role_permission:add") ||
          allowed.has("role_permission:remove") ? (
            <Button type="submit" disabled={assignmentPending}>
              {assignmentPending ? "Saving…" : "Save assignments"}
            </Button>
          ) : null}
        </form>
      ) : (
        <p className="text-muted-foreground">
          Permission assignments are not visible with the current access.
        </p>
      )}
      <RoleDialog
        open={editOpen}
        onClose={() => setEditOpen(false)}
        role={role}
        action={saveRoleAction}
        title={`Edit ${role.name}`}
        mode="edit"
      />
      <ConfirmRoleDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        role={role}
        action={removeRoleAction}
        title={`Delete ${role.name}`}
        submit="Delete role"
        warning="Assigned users or permissions may cause the backend to reject this deletion."
      />
      <ConfirmRoleDialog
        open={defaultOpen}
        onClose={() => setDefaultOpen(false)}
        role={role}
        action={setDefaultRoleAction}
        title={`Make ${role.name} default`}
        submit="Make default"
        warning="New accounts will receive this role unless another role is explicitly selected."
      />
    </section>
  );
}

export function RoleDialog({
  open,
  onClose,
  role,
  action,
  title,
  mode,
}: {
  open: boolean;
  onClose: () => void;
  role?: RoleResponse;
  action: typeof saveRoleAction;
  title: string;
  mode: "create" | "edit";
}) {
  const [state, formAction, pending] = useActionState(
    action,
    EMPTY_ACCESS_STATE,
  );
  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={title}
      variant="sheet"
    >
      <form action={formAction} className="space-y-4">
        {role ? <input type="hidden" name="role_id" value={role.id} /> : null}
        <div>
          <Label htmlFor={`${mode}-role-name`}>Name</Label>
          <Input
            id={`${mode}-role-name`}
            name="name"
            defaultValue={role?.name}
            required
          />
        </div>
        <div>
          <Label htmlFor={`${mode}-role-description`}>Description</Label>
          <textarea
            id={`${mode}-role-description`}
            name="description"
            defaultValue={role?.description}
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
            {pending ? "Saving…" : "Save role"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

function ConfirmRoleDialog({
  open,
  onClose,
  role,
  action,
  title,
  submit,
  warning,
}: {
  open: boolean;
  onClose: () => void;
  role: RoleResponse;
  action: typeof removeRoleAction;
  title: string;
  submit: string;
  warning: string;
}) {
  const [state, formAction, pending] = useActionState(
    action,
    EMPTY_ACCESS_STATE,
  );
  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={title}
      variant="sheet"
    >
      <form action={formAction} className="space-y-4">
        <input type="hidden" name="role_id" value={role.id} />
        <input type="hidden" name="role_name" value={role.name} />
        <p className="bg-muted rounded-xl p-4 text-sm">
          {warning} Enter <strong>{role.name}</strong> to continue.
        </p>
        <Input name="confirmation" aria-label="Confirm role name" required />
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
            {pending ? "Saving…" : submit}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
