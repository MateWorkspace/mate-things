"use client";

import { Search, X } from "lucide-react";
import Form from "next/form";
import Link from "next/link";

import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import type { ApiKeyStatus } from "@/lib/api/api-keys";

interface ApiKeyFiltersProps {
  search?: string;
  status: ApiKeyStatus;
  limit: number;
}

const SELECT_CLASS_NAME =
  "border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none";

export default function ApiKeyFilters({
  search,
  status,
  limit,
}: ApiKeyFiltersProps) {
  return (
    <Form
      action="/admin/api-keys"
      className="grid gap-4 lg:grid-cols-12 lg:items-end"
    >
      <input name="page" type="hidden" value="1" />
      <input name="limit" type="hidden" value={limit} />

      <label className="lg:col-span-6">
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Search
        </span>
        <span className="relative block">
          <Search
            aria-hidden="true"
            className="text-muted-foreground pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2"
          />
          <Input
            className="pl-10"
            defaultValue={search}
            name="search"
            placeholder="Name or username"
            type="search"
          />
        </span>
      </label>

      <label className="lg:col-span-4">
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Status
        </span>
        <select className={SELECT_CLASS_NAME} defaultValue={status} name="status">
          <option value="any">Any status</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </label>

      <div className="flex gap-2 lg:col-span-2">
        <Button className="min-h-11 flex-1" type="submit">
          Apply
        </Button>
        <Link
          aria-label="Clear API key filters"
          className="border-border text-primary hover:bg-highlight/40 focus-visible:ring-focus focus-visible:ring-offset-background inline-flex size-11 shrink-0 items-center justify-center rounded-xl border transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
          href={`/admin/api-keys?limit=${limit}`}
        >
          <X aria-hidden="true" className="size-4" />
        </Link>
      </div>
    </Form>
  );
}
