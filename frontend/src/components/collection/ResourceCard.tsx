import type { ReactNode } from "react";

import Card from "@/components/ui/card";

interface ResourceCardProps {
  actions?: ReactNode;
  children: ReactNode;
  summary?: ReactNode;
  title: string;
}

export default function ResourceCard({
  actions,
  children,
  summary,
  title,
}: ResourceCardProps) {
  return (
    <Card className="flex h-full flex-col gap-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="font-display text-primary text-xl tracking-wide">
            {title}
          </h2>
          {summary ? (
            <p className="text-muted-foreground mt-1 text-sm">{summary}</p>
          ) : null}
        </div>
        {actions ? <div className="shrink-0">{actions}</div> : null}
      </div>
      <div className="text-sm leading-6">{children}</div>
    </Card>
  );
}
