"use client";

interface Tab {
  id: string;
  label: string;
  disabled?: boolean;
}

interface TabsProps {
  tabs: readonly Tab[];
  activeTab: string;
  onChange: (id: string) => void;
  ariaLabel: string;
}

export default function Tabs({
  tabs,
  activeTab,
  onChange,
  ariaLabel,
}: TabsProps) {
  return (
    <div
      role="tablist"
      aria-label={ariaLabel}
      className="border-border flex gap-1 overflow-x-auto border-b"
    >
      {tabs.map((tab) => {
        const isActive = tab.id === activeTab;

        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={isActive}
            disabled={tab.disabled}
            onClick={() => onChange(tab.id)}
            className={`focus-visible:ring-focus focus-visible:ring-offset-background shrink-0 border-b-2 px-3 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 ${
              isActive
                ? "border-primary text-primary"
                : "text-foreground/65 hover:text-foreground border-transparent"
            }`}
          >
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
