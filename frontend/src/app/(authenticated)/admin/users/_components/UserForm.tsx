"use client";

import { useActionState, useId, useState } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { RoleResponse } from "@/lib/api/roles";
import type { UserResponse } from "@/lib/api/users";

import {
  createUserAction,
  deleteUserAction,
  EMPTY_USER_STATE,
  updateUserAction,
} from "../_lib/actions";

interface UserFormProps {
  roles: readonly RoleResponse[];
  user?: UserResponse;
  canEdit?: boolean;
  canDelete?: boolean;
  currentUserId?: string;
}

export default function UserForm({
  roles,
  user,
  canEdit = true,
  canDelete = false,
  currentUserId,
}: UserFormProps) {
  const [editorOpen, setEditorOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  return (
    <div className="flex flex-wrap gap-2">
      {!user || canEdit ? (
        <Button
          type="button"
          variant={user ? "secondary" : "primary"}
          onClick={() => setEditorOpen(true)}
        >
          {user ? "Edit user" : "Create user"}
        </Button>
      ) : null}
      {user && canDelete ? (
        <button
          type="button"
          className="border-critical text-critical focus-visible:ring-critical rounded-xl border px-4 py-2.5 text-sm font-semibold focus-visible:ring-2 focus-visible:outline-none"
          onClick={() => setDeleteOpen(true)}
        >
          Delete user
        </button>
      ) : null}
      <EditorDialog
        key={`${editorOpen}-${user?.id ?? "new"}`}
        open={editorOpen}
        onClose={() => setEditorOpen(false)}
        roles={roles}
        user={user}
      />
      {user ? (
        <DeleteDialog
          key={`${deleteOpen}-${user.id}`}
          open={deleteOpen}
          onClose={() => setDeleteOpen(false)}
          user={user}
          isSelf={currentUserId === user.id}
        />
      ) : null}
    </div>
  );
}

function EditorDialog({
  open,
  onClose,
  roles,
  user,
}: {
  open: boolean;
  onClose: () => void;
  roles: readonly RoleResponse[];
  user?: UserResponse;
}) {
  const [state, action, pending] = useActionState(
    user ? updateUserAction : createUserAction,
    EMPTY_USER_STATE,
  );
  const id = useId();
  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={user ? `Edit ${user.name}` : "Create user"}
      variant="sheet"
    >
      <form action={action} className="space-y-4">
        {user ? <input type="hidden" name="user_id" value={user.id} /> : null}
        <Field
          id={`${id}-name`}
          label="Name"
          name="name"
          value={user?.name}
          error={state.fieldErrors?.name}
        />
        <Field
          id={`${id}-username`}
          label="Username"
          name="username"
          value={user?.username}
          error={state.fieldErrors?.username}
          autoComplete="off"
        />
        <div>
          <Label htmlFor={`${id}-role`}>Role</Label>
          <select
            id={`${id}-role`}
            name="role_id"
            defaultValue={
              user?.role_id ?? roles.find((role) => role.is_default)?.id ?? ""
            }
            required
            className="border-control-border bg-background w-full rounded-xl border px-3.5 py-2.5 text-sm"
          >
            <option value="">Select a role</option>
            {roles.map((role) => (
              <option key={role.id} value={role.id}>
                {role.name}
                {role.is_default ? " (default)" : ""}
              </option>
            ))}
          </select>
        </div>
        <div>
          <Label htmlFor={`${id}-bio`}>Bio</Label>
          <textarea
            id={`${id}-bio`}
            name="bio"
            defaultValue={user?.bio}
            rows={3}
            className="border-control-border bg-background w-full rounded-xl border px-3.5 py-2.5 text-sm"
          />
        </div>
        {!user ? (
          <Field
            id={`${id}-password`}
            label="Initial password"
            name="password"
            type="password"
            error={state.fieldErrors?.password}
            autoComplete="new-password"
          />
        ) : null}
        <ActionMessage state={state} />
        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="secondary"
            disabled={pending}
            onClick={onClose}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Saving…" : user ? "Save changes" : "Create user"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

function DeleteDialog({
  open,
  onClose,
  user,
  isSelf,
}: {
  open: boolean;
  onClose: () => void;
  user: UserResponse;
  isSelf: boolean;
}) {
  const [state, action, pending] = useActionState(
    deleteUserAction,
    EMPTY_USER_STATE,
  );
  const [confirmation, setConfirmation] = useState("");
  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={`Delete ${user.name}`}
      variant="sheet"
    >
      <form action={action} className="space-y-4">
        <input type="hidden" name="user_id" value={user.id} />
        <input type="hidden" name="username" value={user.username} />
        <p className="border-critical/40 bg-critical/5 rounded-xl border p-4 text-sm">
          {isSelf
            ? "This is your signed-in account. Deleting it will end your session. "
            : ""}
          This cannot be undone. Enter <strong>{user.username}</strong> to
          continue.
        </p>
        <Input
          name="confirmation"
          aria-label="Confirm username"
          value={confirmation}
          onChange={(event) => setConfirmation(event.target.value)}
        />
        <ActionMessage state={state} />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <button
            type="submit"
            disabled={pending || confirmation !== user.username}
            className="bg-critical text-background rounded-xl px-4 py-2.5 text-sm font-semibold disabled:opacity-50"
          >
            {pending ? "Deleting…" : "Delete account"}
          </button>
        </div>
      </form>
    </Dialog>
  );
}

function Field({
  id,
  label,
  name,
  value,
  error,
  type = "text",
  autoComplete,
}: {
  id: string;
  label: string;
  name: string;
  value?: string;
  error?: string;
  type?: string;
  autoComplete?: string;
}) {
  return (
    <div>
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        name={name}
        type={type}
        defaultValue={value}
        required
        aria-invalid={Boolean(error)}
        autoComplete={autoComplete}
      />
      {error ? <p className="text-critical mt-1 text-sm">{error}</p> : null}
    </div>
  );
}

function ActionMessage({
  state,
}: {
  state: { status: string; message?: string };
}) {
  return (
    <p
      aria-live="polite"
      className={
        state.status === "error"
          ? "text-critical text-sm"
          : "text-success text-sm"
      }
    >
      {state.message}
    </p>
  );
}
