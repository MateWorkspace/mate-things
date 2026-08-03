"use client";

import { useEffect, useState } from "react";

type LocalDateTimeVariant = "datetime" | "date" | "time";

const FORMAT_OPTIONS: Record<
  LocalDateTimeVariant,
  Intl.DateTimeFormatOptions
> = {
  datetime: { dateStyle: "medium", timeStyle: "short" },
  date: { dateStyle: "medium" },
  time: { timeStyle: "medium" },
};

interface LocalDateTimeProps {
  value: string;
  variant?: LocalDateTimeVariant;
  className?: string;
}

export default function LocalDateTime({
  value,
  variant = "datetime",
  className,
}: LocalDateTimeProps) {
  const [formatted, setFormatted] = useState<string | null>(null);

  useEffect(() => {
    setFormatted(
      new Intl.DateTimeFormat("en", FORMAT_OPTIONS[variant]).format(
        new Date(value),
      ),
    );
  }, [value, variant]);

  return (
    <time dateTime={value} className={className}>
      {formatted ?? "—"}
    </time>
  );
}
