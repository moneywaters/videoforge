import { cn } from "@/lib/utils";
import { ReactNode } from "react";

type Tab = {
  id: string;
  label: string;
};

type TabsProps = {
  tabs: Tab[];
  activeTab: string;
  onChange: (id: string) => void;
  children: ReactNode;
  className?: string;
};

export function Tabs({ tabs, activeTab, onChange, children, className }: TabsProps) {
  return (
    <div className={cn("w-full", className)}>
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8" aria-label="Tabs">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => onChange(tab.id)}
              className={cn(
                "border-b-2 px-1 py-4 text-sm font-medium transition-colors whitespace-nowrap",
                activeTab === tab.id
                  ? "border-brand-600 text-brand-600"
                  : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              )}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>
      <div className="mt-4">{children}</div>
    </div>
  );
}