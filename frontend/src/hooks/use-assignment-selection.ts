import { useState } from "react";

type AssignmentResult = {
  appliedIds: readonly string[];
  failed: ReadonlyArray<{ id: string }>;
  submission?: AssignmentSubmission;
};

export type AssignmentSubmission = {
  values: ReadonlyArray<DesiredValue & { id: string }>;
};

type DesiredValue = {
  checked: boolean;
  revision: number;
};

type AssignmentSelection<State> = {
  handled: State;
  authoritativeKey: string;
  overrides: Map<string, DesiredValue>;
  awaitingConfirmation: Map<string, DesiredValue>;
  revision: number;
  resultGeneration: number;
};

function selectionKey(selected: ReadonlySet<string>): string {
  return JSON.stringify([...selected].sort());
}

export function useAssignmentSelection<State extends AssignmentResult>(
  actionState: State,
  authoritative: ReadonlySet<string>,
  ids: readonly string[],
) {
  const authoritativeKey = selectionKey(authoritative);
  const [storedSelection, setStoredSelection] = useState<
    AssignmentSelection<State>
  >(() => ({
    handled: actionState,
    authoritativeKey,
    overrides: new Map(),
    awaitingConfirmation: new Map(),
    revision: 0,
    resultGeneration: 0,
  }));
  let selection = storedSelection;

  if (
    selection.handled !== actionState ||
    selection.authoritativeKey !== authoritativeKey
  ) {
    const overrides = new Map(selection.overrides);
    const awaitingConfirmation = new Map(selection.awaitingConfirmation);
    let resultGeneration = selection.resultGeneration;

    if (selection.handled !== actionState) {
      resultGeneration += 1;
      const submittedValues = new Map(
        actionState.submission?.values.map(({ id, ...value }) => [id, value]),
      );
      for (const id of [
        ...actionState.appliedIds,
        ...actionState.failed.map(({ id }) => id),
      ]) {
        const desired = submittedValues.get(id);
        if (desired && !overrides.has(id)) {
          overrides.set(id, desired);
        }
      }
      for (const id of actionState.appliedIds) {
        const desired = submittedValues.get(id);
        if (desired) {
          awaitingConfirmation.set(id, desired);
        }
      }
    }

    if (selection.authoritativeKey !== authoritativeKey) {
      for (const [id, confirmedValue] of awaitingConfirmation) {
        if (authoritative.has(id) !== confirmedValue.checked) continue;
        awaitingConfirmation.delete(id);
        const override = overrides.get(id);
        if (override?.revision === confirmedValue.revision) {
          overrides.delete(id);
        }
      }
    }

    selection = {
      ...selection,
      handled: actionState,
      authoritativeKey,
      overrides,
      awaitingConfirmation,
      resultGeneration,
    };
    setStoredSelection(selection);
  }

  return {
    isSelected: (id: string) =>
      selection.overrides.get(id)?.checked ?? authoritative.has(id),
    setSelected: (id: string, checked: boolean) => {
      setStoredSelection((current) => {
        const overrides = new Map(current.overrides);
        const revision = current.revision + 1;
        overrides.set(id, { checked, revision });
        return { ...current, overrides, revision };
      });
    },
    revision: selection.revision,
    resultGeneration: selection.resultGeneration,
    submissionSnapshot: JSON.stringify(
      ids.map((id) => ({
        id,
        revision: selection.overrides.get(id)?.revision ?? 0,
      })),
    ),
  };
}

export function assignmentSubmission(
  data: FormData,
  selectedField: string,
): AssignmentSubmission {
  const selectedIds = new Set(
    data
      .getAll(selectedField)
      .filter((id): id is string => typeof id === "string"),
  );
  let snapshot: unknown = [];
  try {
    snapshot = JSON.parse(String(data.get("assignment_snapshot") ?? "[]"));
  } catch {
    // A malformed client-only snapshot cannot affect the server operation.
  }
  return {
    values: Array.isArray(snapshot)
      ? snapshot.flatMap((entry): Array<DesiredValue & { id: string }> => {
          if (!entry || typeof entry !== "object") return [];
          const id = "id" in entry ? entry.id : undefined;
          const revision = "revision" in entry ? Number(entry.revision) : 0;
          if (typeof id !== "string") return [];
          return [
            {
              id,
              checked: selectedIds.has(id),
              revision: Number.isFinite(revision) ? revision : 0,
            },
          ];
        })
      : [],
  };
}
