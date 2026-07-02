"use client";

import { Check } from "lucide-react";
import { Button } from "@/components/ui";
import { BOOKING_TOPICS } from "@/constants";
import { useTranslation } from "@/hooks";

export interface TopicStepProps {
  selectedTopics: string[];
  onToggleTopic: (topicId: string) => void;
  onNext: () => void;
}

export function TopicStep({ selectedTopics, onToggleTopic, onNext }: TopicStepProps) {
  const { t } = useTranslation();

  return (
    <div className="flex h-full flex-col">
      <div className="mb-8">
        <h1 className="mb-3 text-3xl font-bold text-foreground">What brings you here today?</h1>
        <p className="max-w-2xl text-sm text-muted-foreground">
          Please select the topics or symptoms you&apos;d like to discuss with an expert. You can choose multiple.
        </p>
      </div>

      <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
        {BOOKING_TOPICS.map((topic) => {
          const isSelected = selectedTopics.includes(topic.id);
          const Icon = topic.icon;

          return (
            <button
              key={topic.id}
              type="button"
              onClick={() => onToggleTopic(topic.id)}
              className={`relative flex h-36 flex-col items-center justify-center rounded-2xl border p-6 transition-all ${
                isSelected
                  ? "border-primary bg-primary-soft/50 shadow-elevated"
                  : "border-border bg-background hover:border-primary/40 hover:shadow-card"
              }`}
            >
              {isSelected && (
                <div className="absolute right-3 top-3 text-primary">
                  <Check className="h-4 w-4 stroke-[3]" />
                </div>
              )}
              <Icon className={`mb-4 h-8 w-8 ${isSelected ? "text-primary" : "text-muted-foreground"}`} strokeWidth={1.5} />
              <span className={`text-center text-sm font-semibold ${isSelected ? "text-primary-soft-text" : "text-foreground"}`}>
                {t(topic.labelKey, topic.label)}
              </span>
            </button>
          );
        })}
      </div>

      <div className="mt-auto flex justify-end border-t border-border pt-6">
        <Button onClick={onNext} disabled={selectedTopics.length === 0}>
          Next
          <Check className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
