"use client";

import { useState, type ReactNode } from "react";
import { SlidersHorizontal } from "lucide-react";

import Button from "@/components/ui/button";

import FilterDrawer from "./FilterDrawer";

interface CollectionToolbarProps {
  children: ReactNode;
  filterTitle?: string;
}

export default function CollectionToolbar({
  children,
  filterTitle = "Filters",
}: CollectionToolbarProps) {
  const [drawerOpen, setDrawerOpen] = useState(false);

  return (
    <div className="border-border bg-surface rounded-2xl border p-3 sm:p-4">
      {drawerOpen ? null : <div className="hidden lg:block">{children}</div>}
      <Button
        aria-haspopup="dialog"
        aria-expanded={drawerOpen}
        className="w-full gap-2 lg:hidden"
        type="button"
        variant="secondary"
        onClick={() => setDrawerOpen(true)}
      >
        <SlidersHorizontal aria-hidden="true" className="size-4" />
        Open filters
      </Button>
      <FilterDrawer
        open={drawerOpen}
        title={filterTitle}
        onClose={() => setDrawerOpen(false)}
      >
        {children}
      </FilterDrawer>
    </div>
  );
}
