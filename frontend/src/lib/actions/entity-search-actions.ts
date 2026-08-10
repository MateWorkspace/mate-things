"use server";

import { listActions } from "@/lib/api/actions";
import { listNodeClasses } from "@/lib/api/node-classes";
import { listNodes } from "@/lib/api/nodes";
import { listRoles } from "@/lib/api/roles";
import { listUsers } from "@/lib/api/users";
import { requirePermission } from "@/lib/session";

import {
  parseSearchOptionsRequest,
  toSearchOptionsPage,
  type SearchOptionsPage,
  type SearchOptionsRequest,
} from "./search-options";

function toListQuery(request: SearchOptionsRequest) {
  return {
    search: request.query || undefined,
    page: request.page,
    ...(request.limit === undefined ? {} : { limit: request.limit }),
  };
}

export async function searchNodesAction(
  input: unknown,
): Promise<SearchOptionsPage> {
  const request = parseSearchOptionsRequest(input);
  await requirePermission("node:get");
  const response = await listNodes(toListQuery(request));
  return toSearchOptionsPage(response, (node) => ({
    value: node.id,
    label: node.name,
    description: node.device_id,
  }));
}

export async function searchNodeDeviceIdsAction(
  input: unknown,
): Promise<SearchOptionsPage> {
  const request = parseSearchOptionsRequest(input);
  await requirePermission("node:get");
  const response = await listNodes(toListQuery(request));
  return toSearchOptionsPage(response, (node) => ({
    value: node.device_id,
    label: node.name || node.device_id,
    description: node.device_id,
  }));
}

export async function searchActionsAction(
  input: unknown,
): Promise<SearchOptionsPage> {
  const request = parseSearchOptionsRequest(input);
  await requirePermission("action:get");
  const response = await listActions(toListQuery(request));
  return toSearchOptionsPage(response, (action) => ({
    value: action.id,
    label: action.name,
    description: action.description,
  }));
}

export async function searchNodeClassesAction(
  input: unknown,
): Promise<SearchOptionsPage> {
  const request = parseSearchOptionsRequest(input);
  await requirePermission("node_class:get");
  const response = await listNodeClasses(toListQuery(request));
  return toSearchOptionsPage(response, (nodeClass) => ({
    value: nodeClass.id,
    label: nodeClass.name,
    description: nodeClass.description,
  }));
}

export async function searchRolesAction(
  input: unknown,
): Promise<SearchOptionsPage> {
  const request = parseSearchOptionsRequest(input);
  await requirePermission("role:get");
  const response = await listRoles(toListQuery(request));
  return toSearchOptionsPage(response, (role) => ({
    value: role.id,
    label: role.name,
    description: role.description,
  }));
}

export async function searchUsersAction(
  input: unknown,
): Promise<SearchOptionsPage> {
  const request = parseSearchOptionsRequest(input);
  await requirePermission("user:get");
  const response = await listUsers(toListQuery(request));
  return toSearchOptionsPage(response, (user) => ({
    value: user.id,
    label: `${user.name} (@${user.username})`,
  }));
}
