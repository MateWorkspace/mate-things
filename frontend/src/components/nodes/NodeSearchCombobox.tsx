"use client";

import { ChevronLeft, ChevronRight, Search, X } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";

import Input from "@/components/ui/input";
import { searchNodesAction } from "@/lib/actions/entity-search-actions";
import type { SearchOption } from "@/lib/actions/search-options";

const RESULTS_PER_PAGE = 6;
const DEBOUNCE_MS = 300;

interface NodeSearchComboboxProps {
  name: string;
  defaultNodeId?: string;
  defaultNodeName?: string;
  label?: string;
  placeholder?: string;
}

export default function NodeSearchCombobox({
  name,
  defaultNodeId,
  defaultNodeName,
  label = "Node",
  placeholder = "Any node",
}: NodeSearchComboboxProps) {
  const id = useId();
  const containerRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);
  const [nodes, setNodes] = useState<SearchOption[]>([]);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(false);
  const [selectedId, setSelectedId] = useState(defaultNodeId ?? "");
  const [selectedName, setSelectedName] = useState(defaultNodeName ?? "");

  useEffect(() => {
    if (!open) return undefined;

    const timeout = setTimeout(() => {
      let cancelled = false;
      setLoading(true);
      searchNodesAction({ query, page, limit: RESULTS_PER_PAGE })
        .then((result) => {
          if (cancelled) return;
          setNodes(result.items);
          setTotalPages(result.totalPages);
        })
        .catch(() => {
          if (cancelled) return;
          setNodes([]);
          setTotalPages(1);
        })
        .finally(() => {
          if (!cancelled) setLoading(false);
        });
      return () => {
        cancelled = true;
      };
    }, DEBOUNCE_MS);

    return () => clearTimeout(timeout);
  }, [open, query, page]);

  useEffect(() => {
    if (!open) return undefined;

    function handlePointerDown(event: MouseEvent) {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [open]);

  function select(node: SearchOption | null) {
    setSelectedId(node?.value ?? "");
    setSelectedName(node?.label ?? "");
    setOpen(false);
  }

  return (
    <div ref={containerRef} className="relative">
      <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
        {label}
      </span>
      <input type="hidden" name={name} value={selectedId} />
      <button
        type="button"
        data-field-name={name}
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => {
          setOpen((current) => !current);
          setPage(1);
        }}
        className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus flex min-h-11 w-full items-center justify-between gap-2 rounded-xl border px-3.5 py-2.5 text-left text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
      >
        <span className={selectedName ? "" : "text-foreground/40"}>
          {selectedName || placeholder}
        </span>
        <ChevronRight
          aria-hidden="true"
          className={`size-4 shrink-0 transition-transform ${open ? "rotate-90" : ""}`}
        />
      </button>

      {open ? (
        <div className="border-border bg-surface absolute z-20 mt-1.5 w-full min-w-64 rounded-xl border p-2 shadow-lg">
          <div className="relative">
            <Search
              aria-hidden="true"
              className="text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2"
            />
            <Input
              autoFocus
              className="pl-9"
              placeholder="Search nodes"
              value={query}
              onChange={(event) => {
                setQuery(event.target.value);
                setPage(1);
              }}
              aria-label="Search nodes"
            />
          </div>

          <ul
            className="mt-2 max-h-56 space-y-0.5 overflow-y-auto"
            role="listbox"
          >
            <li>
              <button
                type="button"
                onClick={() => select(null)}
                className="hover:bg-highlight/40 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors"
              >
                <X
                  aria-hidden="true"
                  className="text-muted-foreground size-3.5"
                />
                Any node
              </button>
            </li>
            {loading ? (
              <li className="text-muted-foreground px-2.5 py-2 text-sm">
                Searching…
              </li>
            ) : nodes.length ? (
              nodes.map((node) => (
                <li key={node.value}>
                  <button
                    type="button"
                    role="option"
                    aria-selected={node.value === selectedId}
                    onClick={() => select(node)}
                    className={`hover:bg-highlight/40 w-full rounded-lg px-2.5 py-2 text-left text-sm transition-colors ${
                      node.value === selectedId
                        ? "bg-highlight/40 font-semibold"
                        : ""
                    }`}
                  >
                    {node.label}
                  </button>
                </li>
              ))
            ) : (
              <li className="text-muted-foreground px-2.5 py-2 text-sm">
                No nodes match.
              </li>
            )}
          </ul>

          <div className="border-border mt-2 flex items-center justify-between border-t pt-2">
            <span
              id={`${id}-page-info`}
              className="text-muted-foreground text-xs"
            >
              Page {page} of {totalPages}
            </span>
            <div className="flex gap-1">
              <button
                type="button"
                aria-label="Previous nodes"
                disabled={page <= 1 || loading}
                onClick={() => setPage((current) => Math.max(1, current - 1))}
                className="hover:bg-highlight/40 flex size-7 items-center justify-center rounded-lg transition-colors disabled:opacity-40"
              >
                <ChevronLeft aria-hidden="true" className="size-4" />
              </button>
              <button
                type="button"
                aria-label="Next nodes"
                disabled={page >= totalPages || loading}
                onClick={() =>
                  setPage((current) => Math.min(totalPages, current + 1))
                }
                className="hover:bg-highlight/40 flex size-7 items-center justify-center rounded-lg transition-colors disabled:opacity-40"
              >
                <ChevronRight aria-hidden="true" className="size-4" />
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
