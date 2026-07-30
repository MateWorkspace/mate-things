"use client";

import { useState } from "react";

import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { toDatetimeLocal } from "@/lib/record-filters";

interface TimeRangeFilterProps {
  start?: string;
  end?: string;
  startName?: string;
  endName?: string;
}

export default function TimeRangeFilter({
  start,
  end,
  startName = "start",
  endName = "end",
}: TimeRangeFilterProps) {
  const [startValue, setStartValue] = useState(toDatetimeLocal(start));
  const [endValue, setEndValue] = useState(toDatetimeLocal(end));

  return (
    <>
      <div>
        <Label htmlFor={`${startName}-filter`}>From</Label>
        <Input
          id={`${startName}-filter`}
          type="datetime-local"
          value={startValue}
          onChange={(event) => setStartValue(event.target.value)}
        />
        <input
          type="hidden"
          name={startName}
          value={startValue ? new Date(startValue).toISOString() : ""}
        />
      </div>
      <div>
        <Label htmlFor={`${endName}-filter`}>To</Label>
        <Input
          id={`${endName}-filter`}
          type="datetime-local"
          value={endValue}
          onChange={(event) => setEndValue(event.target.value)}
        />
        <input
          type="hidden"
          name={endName}
          value={endValue ? new Date(endValue).toISOString() : ""}
        />
      </div>
    </>
  );
}
