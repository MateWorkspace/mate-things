import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, expect, it, vi } from "vitest";

import type { PayloadSchemaResponse } from "@/lib/api/payload-schemas";

import PayloadSchemaForm from "./PayloadSchemaForm";

const mocks = vi.hoisted(() => ({
  createPayloadSchemaAction: vi.fn(),
  deletePayloadSchemaAction: vi.fn(),
  refresh: vi.fn(),
  updatePayloadSchemaAction: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: mocks.refresh }),
}));

vi.mock("../_lib/actions", () => ({
  createPayloadSchemaAction: mocks.createPayloadSchemaAction,
  deletePayloadSchemaAction: mocks.deletePayloadSchemaAction,
  updatePayloadSchemaAction: mocks.updatePayloadSchemaAction,
}));

const SCHEMA: PayloadSchemaResponse = {
  id: "schema-1",
  name: "temperature_reading",
  version: 1,
  definition: {
    type: "object",
    properties: { temperature: { type: "float" } },
  },
  valid_from: "2026-08-10T00:00:00Z",
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

beforeEach(() => {
  vi.resetAllMocks();
  mocks.updatePayloadSchemaAction.mockImplementation(
    async (expectedId: unknown) =>
      expectedId === SCHEMA.id
        ? {
            status: "success",
            title: "Schema updated",
            message: "Schema saved.",
          }
        : {
            status: "error",
            title: "Cannot update schema",
            message: "A trusted update identity is required.",
          },
  );
});

afterEach(cleanup);

it("binds the authoritative schema id to the update action", async () => {
  const user = userEvent.setup();
  render(<PayloadSchemaForm schema={SCHEMA} />);

  await user.click(screen.getByRole("button", { name: "Edit schema" }));
  await user.click(screen.getByRole("button", { name: "Save schema" }));

  await waitFor(() =>
    expect(mocks.updatePayloadSchemaAction).toHaveBeenCalled(),
  );
  expect(mocks.updatePayloadSchemaAction.mock.calls[0]?.[0]).toBe(SCHEMA.id);
});
