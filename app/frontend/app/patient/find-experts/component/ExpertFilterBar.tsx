"use client";

import { Search } from "lucide-react";
import type { Specialization } from "@/types";

export interface ExpertFilterBarProps {
  query: string;
  onQueryChange: (query: string) => void;
  specializationId: string;
  onSpecializationChange: (specializationId: string) => void;
  specializations: Specialization[];
}

export function ExpertFilterBar({
  query,
  onQueryChange,
  specializationId,
  onSpecializationChange,
  specializations,
}: ExpertFilterBarProps) {
  return (
    <div className="mb-6 flex flex-col gap-3 sm:flex-row">
      <div className="relative flex-1">
        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-muted-foreground">
          <Search className="h-4 w-4" />
        </div>
        <input
          type="text"
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
          placeholder="Search by expert name or specialization..."
          className="w-full rounded-xl border border-border py-2.5 pl-10 pr-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
        />
      </div>
      <select
        value={specializationId}
        onChange={(event) => onSpecializationChange(event.target.value)}
        className="rounded-xl border border-border px-3 py-2.5 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
      >
        <option value="">All specializations</option>
        {specializations.map((spec) => (
          <option key={spec.specId} value={spec.specId}>
            {spec.name}
          </option>
        ))}
      </select>
    </div>
  );
}
