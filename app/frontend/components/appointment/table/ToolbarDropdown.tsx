"use client";

import type { ReactNode } from "react";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { ChevronDown } from "lucide-react";

export interface ToolbarDropdownProps {
  label: string;
  children: ReactNode;
}

/** Compact "Screen Options" / "Help" style dropdown for the page header. */
export function ToolbarDropdown({ label, children }: ToolbarDropdownProps) {
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild>
        <button
          type="button"
          className="inline-flex h-7 items-center gap-1 rounded-md border border-border bg-background px-2.5 text-xs font-medium text-muted-foreground hover:bg-surface hover:text-foreground data-[state=open]:bg-surface"
        >
          {label}
          <ChevronDown className="h-3 w-3" aria-hidden="true" />
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={6}
          className="z-50 min-w-[220px] max-w-[300px] rounded-xl border border-border bg-background p-1.5 shadow-elevated"
        >
          {children}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}

export function ToolbarDropdownCheckbox({
  checked,
  onCheckedChange,
  children,
}: {
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
  children: ReactNode;
}) {
  return (
    <DropdownMenu.CheckboxItem
      checked={checked}
      onCheckedChange={onCheckedChange}
      onSelect={(event) => event.preventDefault()}
      className="flex h-9 cursor-pointer items-center gap-2 rounded-lg px-3 text-sm text-foreground outline-none data-[highlighted]:bg-surface"
    >
      <input type="checkbox" readOnly checked={checked} tabIndex={-1} className="h-3.5 w-3.5 accent-primary" />
      {children}
    </DropdownMenu.CheckboxItem>
  );
}

export function ToolbarDropdownLabel({ children }: { children: ReactNode }) {
  return (
    <DropdownMenu.Label className="px-3 py-1.5 text-xs font-semibold text-muted-foreground">
      {children}
    </DropdownMenu.Label>
  );
}
