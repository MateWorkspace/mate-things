"use client";

import { Plus, X } from "lucide-react";

import Button from "@/components/ui/button";
import Input from "@/components/ui/input";

export default function StringListEditor({
  values,
  onChange,
  placeholder,
}: {
  values: readonly string[];
  onChange: (values: string[]) => void;
  placeholder?: string;
}) {
  return (
    <div className="space-y-2">
      {values.map((value, index) => (
        <div key={index} className="flex items-center gap-2">
          <Input
            value={value}
            placeholder={placeholder}
            onChange={(event) => {
              const next = [...values];
              next[index] = event.target.value;
              onChange(next);
            }}
          />
          <Button
            type="button"
            variant="secondary"
            aria-label="Remove option"
            onClick={() => onChange(values.filter((_, i) => i !== index))}
          >
            <X aria-hidden="true" className="size-4" />
          </Button>
        </div>
      ))}
      <Button
        type="button"
        variant="secondary"
        onClick={() => onChange([...values, ""])}
      >
        <Plus aria-hidden="true" className="mr-1.5 size-4" />
        Add option
      </Button>
    </div>
  );
}
