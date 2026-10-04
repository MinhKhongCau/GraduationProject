import { Card } from "@/components/ui";
import type { AssessmentQuestion } from "@/types";

export interface QuestionStepProps {
  question: AssessmentQuestion;
  selectedOptionId?: string;
  onAnswer: (optionId: string) => void;
}

export function QuestionStep({ question, selectedOptionId, onAnswer }: QuestionStepProps) {
  return (
    <Card className="p-6">
      <h2 className="mb-5 text-lg font-semibold text-foreground">{question.content}</h2>
      <div className="space-y-2">
        {question.options
          .slice()
          .sort((a, b) => a.orderIndex - b.orderIndex)
          .map((option) => (
            <button
              key={option.slug}
              type="button"
              onClick={() => onAnswer(option.slug)}
              aria-pressed={selectedOptionId === option.slug}
              className={`w-full rounded-xl border px-4 py-3 text-left text-sm font-medium transition-colors ${
                selectedOptionId === option.slug
                  ? "border-primary bg-primary-soft text-primary-soft-text ring-1 ring-primary"
                  : "border-border-strong text-foreground hover:border-primary/40 hover:bg-surface"
              }`}
            >
              {option.label}
            </button>
          ))}
      </div>
    </Card>
  );
}
