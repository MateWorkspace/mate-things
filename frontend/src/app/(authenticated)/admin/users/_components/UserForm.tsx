"use client";

import { useActionState, useEffect, useId, useRef, useState } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import type { RoleResponse } from "@/lib/api/roles";
import type { UserResponse } from "@/lib/api/users";

import {
  createUserAction,
  deleteUserAction,
  updateUserAction,
} from "../_lib/actions";
import { EMPTY_USER_STATE, type UserActionState } from "../_lib/state";

function useActionToast(state: UserActionState): void {
  const toast = useToast();
  const lastShown = useRef<UserActionState | null>(null);
  useEffect(() => {
    if (state.status === "idle" || state === lastShown.current) return;
    lastShown.current = state;
    if (state.status === "error") {
      toast.error(state.title ?? "Something went wrong", state.message ?? "");
    } else {
      toast.success(state.title ?? "Success", state.message ?? "");
    }
  }, [state, toast]);
}

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
  const [editorGeneration, setEditorGeneration] = useState(0);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteGeneration, setDeleteGeneration] = useState(0);
  return (
    <div className="flex flex-wrap gap-2">
      {!user || canEdit ? (
        <Button
          type="button"
          variant={user ? "secondary" : "primary"}
          onClick={() => {
            setEditorGeneration((generation) => generation + 1);
            setEditorOpen(true);
          }}
        >
          {user ? "Edit user" : "Create user"}
        </Button>
      ) : null}
      {user && canDelete ? (
        <Button
          type="button"
          variant="critical"
          onClick={() => {
            setDeleteGeneration((generation) => generation + 1);
            setDeleteOpen(true);
          }}
        >
          Delete user
        </Button>
      ) : null}
      {/* Keyed on a generation counter bumped only when opening, not on
          editorOpen/deleteOpen themselves - remounting on close would
          unmount a still-open native <dialog>, dropping it from the
          browser's top layer without a clean close() call. */}
      <EditorDialog
        key={`${editorGeneration}-${user?.id ?? "new"}`}
        open={editorOpen}
        onClose={() => setEditorOpen(false)}
        roles={roles}
        user={user}
      />
      {user ? (
        <DeleteDialog
          key={`${deleteGeneration}-${user.id}`}
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
  useActionToast(state);
  useEffect(() => {
    if (state.status === "success") {
      onClose();
    }
  }, [state, onClose]);
  const id = useId();

  // Controlled, not defaultValue: React resets a <form action={...}> to its
  // defaults after every submission (success or rejected), which would wipe
  // whatever the user typed on a validation/conflict error. Driving these
  // from state survives that reset.
  const [name, setName] = useState(user?.name ?? "");
  const [username, setUsername] = useState(user?.username ?? "");
  const [bio, setBio] = useState(user?.bio ?? "");
  const [password, setPassword] = useState("");
  const [roleId, setRoleId] = useState(
    user?.role_id ?? roles.find((role) => role.is_default)?.id ?? "",
  );

  return (
    <Dialog
      open={open}
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
          value={name}
          onChange={setName}
          error={state.fieldErrors?.name}
        />
        <Field
          id={`${id}-username`}
          label="Username"
          name="username"
          value={username}
          onChange={setUsername}
          error={state.fieldErrors?.username}
          autoComplete="off"
        />
        <div>
          <Label htmlFor={`${id}-role`}>Role</Label>
          <select
            id={`${id}-role`}
            name="role_id"
            value={roleId}
            onChange={(event) => setRoleId(event.target.value)}
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
            value={bio}
            onChange={(event) => setBio(event.target.value)}
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
            value={password}
            onChange={setPassword}
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
  useActionToast(state);
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
  onChange,
  error,
  type = "text",
  autoComplete,
}: {
  id: string;
  label: string;
  name: string;
  value: string;
  onChange: (value: string) => void;
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
        value={value}
        onChange={(event) => onChange(event.target.value)}
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
