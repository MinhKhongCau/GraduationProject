"use client";

import { Search } from "lucide-react";
import { Input, Select } from "@/components/ui";
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
        <Input
          type="text"
          aria-label="Search experts"
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
          placeholder="Search by expert name or specialization..."
          className="pl-10"
        />
      </div>
      <Select
        aria-label="Filter by specialization"
        value={specializationId}
        onChange={(event) => onSpecializationChange(event.target.value)}
        className="sm:w-64"
      >
        <option value="">All specializations</option>
        {specializations.map((spec) => (
          <option key={spec.specId} value={spec.specId}>
            {spec.name}
          </option>
        ))}
      </Select>
    </div>
  );
}
