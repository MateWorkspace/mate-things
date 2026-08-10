import Link from "next/link";

import { EmptyState } from "@/components/ui/states";

interface NodeCollectionEmptyStateProps {
  firmwareId?: string;
  limit: number;
  nodeClassId?: string;
  search?: string;
  totalItems: number;
}

function StateLink({ children, href }: { children: string; href: string }) {
  return (
    <Link
      className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background inline-flex min-h-11 items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
      href={href}
    >
      {children}
    </Link>
  );
}

export default function NodeCollectionEmptyState({
  firmwareId,
  limit,
  nodeClassId,
  search,
  totalItems,
}: NodeCollectionEmptyStateProps) {
  const hasActiveFilters = Boolean(search || nodeClassId || firmwareId);
  const firstPageHref = `/nodes?limit=${limit}`;

  if (totalItems === 0 && !hasActiveFilters) {
    return (
      <EmptyState
        title="No nodes registered"
        description="Nodes self-register through MQTT when devices connect. There is no manual create action."
      />
    );
  }

  if (hasActiveFilters) {
    return (
      <EmptyState
        title="No matching nodes"
        description="No nodes match the current search and filters."
        action={<StateLink href={firstPageHref}>Clear filters</StateLink>}
      />
    );
  }

  return (
    <EmptyState
      title="No nodes on this page"
      description="This page no longer contains results. Return to the first page to continue browsing."
      action={<StateLink href={firstPageHref}>Return to first page</StateLink>}
    />
  );
}
