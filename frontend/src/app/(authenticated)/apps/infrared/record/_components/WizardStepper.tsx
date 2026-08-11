import { Check } from "lucide-react";

export default function WizardStepper({
  steps,
  currentIndex,
}: {
  steps: readonly string[];
  currentIndex: number;
}) {
  return (
    <ol className="flex flex-wrap gap-x-6 gap-y-3 text-sm">
      {steps.map((label, index) => {
        const done = index < currentIndex;
        const active = index === currentIndex;
        return (
          <li key={label} className="flex items-center gap-2">
            <span
              className={`flex size-6 shrink-0 items-center justify-center rounded-full border text-xs font-semibold ${
                done
                  ? "border-success bg-success text-surface"
                  : active
                    ? "border-primary text-primary"
                    : "border-border text-muted-foreground"
              }`}
            >
              {done ? (
                <Check aria-hidden="true" className="size-3.5" />
              ) : (
                index + 1
              )}
            </span>
            <span
              className={
                active ? "text-foreground font-semibold" : "text-foreground/70"
              }
            >
              {label}
            </span>
          </li>
        );
      })}
    </ol>
  );
}
