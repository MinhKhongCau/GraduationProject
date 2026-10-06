"use client";

import { useState } from "react";
import { Plus, X } from "lucide-react";
import { Card, CardHeader, Button, Select } from "@/components/ui";
import type { Specialization } from "@/types";

export interface MySpecializationsCardProps {
  /** Specializations the expert currently lists (may include ones an admin has since deactivated). */
  current: Specialization[];
  /** Active specializations an expert can add (public list). */
  available: Specialization[];
  onAdd: (specId: string) => void;
  onRemove: (specId: string) => void;
  isPending: boolean;
}

export function MySpecializationsCard({ current, available, onAdd, onRemove, isPending }: MySpecializationsCardProps) {
  const [selectedSpecId, setSelectedSpecId] = useState("");
  const currentIds = new Set(current.map((spec) => spec.specId));
  const addable = available.filter((spec) => !currentIds.has(spec.specId));

  function handleAdd() {
    if (!selectedSpecId) return;
    onAdd(selectedSpecId);
    setSelectedSpecId("");
  }

  return (
    <Card>
      <CardHeader title="My specializations" />
      <div className="space-y-5 p-5 sm:p-6">
        <p className="text-sm text-muted-foreground">
          Patients find you by these specializations when booking.
        </p>

        {current.length === 0 ? (
          <p className="text-sm text-muted-foreground">You haven&apos;t added any specializations yet.</p>
        ) : (
          <ul className="flex flex-wrap gap-2">
            {current.map((spec) => (
              <li
                key={spec.specId}
                className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-primary/20 bg-primary-soft pl-3 pr-1 text-xs font-medium text-primary-soft-text"
              >
                {spec.name}
                {!spec.isActive && <span className="text-muted-foreground">(inactive)</span>}
                <button
                  type="button"
                  onClick={() => onRemove(spec.specId)}
                  disabled={isPending}
                  aria-label={`Remove ${spec.name}`}
                  className="inline-flex h-6 w-6 items-center justify-center rounded-md transition-colors hover:bg-primary/10 disabled:opacity-50"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              </li>
            ))}
          </ul>
        )}

        <div className="flex flex-col gap-2 border-t border-border pt-5 sm:flex-row">
          <Select
            aria-label="Specialization to add"
            value={selectedSpecId}
            onChange={(e) => setSelectedSpecId(e.target.value)}
            disabled={addable.length === 0}
            className="sm:max-w-xs"
          >
            <option value="">{addable.length === 0 ? "No more specializations to add" : "Choose a specialization..."}</option>
            {addable.map((spec) => (
              <option key={spec.specId} value={spec.specId}>
                {spec.name}
              </option>
            ))}
          </Select>
          <Button type="button" onClick={handleAdd} disabled={!selectedSpecId || isPending}>
            <Plus className="h-4 w-4" /> Add specialization
          </Button>
        </div>
      </div>
    </Card>
  );
}
