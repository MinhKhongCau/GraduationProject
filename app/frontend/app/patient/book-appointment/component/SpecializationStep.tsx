"use client";

import { Check, Stethoscope } from "lucide-react";
import { Button, PageHeader, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { specializationApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { Specialization } from "@/types";

export interface SpecializationStepProps {
  selectedSpecialization: Specialization | null;
  onSelectSpecialization: (specialization: Specialization) => void;
  onNext: () => void;
}

export function SpecializationStep({ selectedSpecialization, onSelectSpecialization, onNext }: SpecializationStepProps) {
  const { data: specializations = [], isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.specializations(),
    queryFn: () => specializationApi.getAllSpecializations(),
  });
  const activeSpecializations = specializations.filter((spec) => spec.isActive);

  return (
    <div className="flex h-full flex-col">
      <PageHeader
        title="Choose a specialization"
        description="Pick the area you'd like support with. We'll show experts who practise it."
      />

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : activeSpecializations.length === 0 ? (
        <p className="text-sm text-muted-foreground">No specializations are available right now.</p>
      ) : (
        <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {activeSpecializations.map((spec) => {
            const isSelected = selectedSpecialization?.specId === spec.specId;
            return (
              <button
                key={spec.specId}
                type="button"
                onClick={() => onSelectSpecialization(spec)}
                aria-pressed={isSelected}
                className={`relative flex flex-col gap-2 rounded-xl border p-5 text-left transition-all ${
                  isSelected
                    ? "border-primary bg-primary-soft ring-1 ring-primary"
                    : "border-border-strong bg-background hover:border-primary/40 hover:shadow-card"
                }`}
              >
                {isSelected && (
                  <div className="absolute right-3 top-3 text-primary">
                    <Check className="h-4 w-4 stroke-[3]" />
                  </div>
                )}
                <Stethoscope className={`h-6 w-6 ${isSelected ? "text-primary" : "text-muted-foreground"}`} strokeWidth={1.5} />
                <span className={`text-sm font-semibold ${isSelected ? "text-primary-soft-text" : "text-foreground"}`}>
                  {spec.name}
                </span>
                {spec.description && <span className="line-clamp-2 text-xs text-muted-foreground">{spec.description}</span>}
                {spec.symptoms.length > 0 && (
                  <span className="line-clamp-1 text-xs text-muted-foreground">{spec.symptoms.join(" · ")}</span>
                )}
              </button>
            );
          })}
        </div>
      )}

      <div className="mt-auto flex justify-end border-t border-border pt-6">
        <Button onClick={onNext} disabled={!selectedSpecialization}>
          Next
          <Check className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
