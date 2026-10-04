"use client";

import Image from "next/image";
import { Check } from "lucide-react";
import { Button, PageHeader, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { expertApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { ExpertProfile } from "@/types";

export interface ExpertStepProps {
  selectedExpert: ExpertProfile | null;
  onSelectExpert: (expert: ExpertProfile) => void;
  onBack: () => void;
  onNext: () => void;
}

export function ExpertStep({ selectedExpert, onSelectExpert, onBack, onNext }: ExpertStepProps) {
  const { data: experts = [], isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.experts(),
    queryFn: () => expertApi.getAllExperts(),
  });

  return (
    <div className="flex h-full flex-col">
      <PageHeader title="Choose your expert" />

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : (
        <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {experts.map((expert) => {
            const isSelected = selectedExpert?.accountId === expert.accountId;
            return (
              <button
                key={expert.expertId}
                type="button"
                onClick={() => onSelectExpert(expert)}
                aria-pressed={isSelected}
                className={`relative flex items-center gap-3 rounded-xl border bg-background p-4 text-left transition-all ${
                  isSelected ? "border-primary bg-primary-soft ring-1 ring-primary" : "border-border-strong hover:border-primary/40 hover:shadow-card"
                }`}
              >
                {isSelected && (
                  <div className="absolute right-3 top-3 text-primary">
                    <Check className="h-4 w-4 stroke-[3]" />
                  </div>
                )}
                <Image
                  src={expert.avatarUrl || "/images/auth-bg.png"}
                  alt={expert.fullName}
                  width={44}
                  height={44}
                  unoptimized
                  className="h-11 w-11 rounded-full object-cover"
                />
                <div>
                  <p className="text-sm font-semibold text-foreground">{expert.fullName}</p>
                  <p className="text-xs text-muted-foreground">
                    {expert.specializations.map((spec) => spec.name).join(", ") || "General counseling"}
                  </p>
                </div>
              </button>
            );
          })}
        </div>
      )}

      <div className="mt-auto flex justify-between border-t border-border pt-6">
        <Button variant="outline" onClick={onBack}>
          Back
        </Button>
        <Button onClick={onNext} disabled={!selectedExpert}>
          Next
          <Check className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
