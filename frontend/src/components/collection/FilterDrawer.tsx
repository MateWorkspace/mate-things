"use client";

import type { ReactNode } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";

interface FilterDrawerProps {
  children: ReactNode;
  open: boolean;
  onClose: () => void;
  title: string;
}

export default function FilterDrawer({
  children,
  onClose,
  open,
  title,
}: FilterDrawerProps) {
  return (
    <Dialog open={open} onClose={onClose} title={title} variant="sheet">
      <div className="space-y-4">{children}</div>
      <div className="border-border mt-6 flex justify-end border-t pt-4">
        <Button type="button" variant="secondary" onClick={onClose}>
          Close filters
        </Button>
      </div>
    </Dialog>
  );
}
