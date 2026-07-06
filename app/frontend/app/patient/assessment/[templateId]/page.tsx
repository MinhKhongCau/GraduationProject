"use client";

import { use, useState } from "react";
import { QuestionStep } from "./component/QuestionStep";
import { ProgressBar } from "./component/ProgressBar";
import { ResultSummary } from "./component/ResultSummary";
import { Button, Spinner } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { assessmentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";
import type { AnswerSubmit } from "@/types";

export default function AssessmentDetailPage({ params }: { params: Promise<{ templateId: string }> }) {
  const { templateId } = use(params);
  const { user } = useAuthContext();

  const [currentIndex, setCurrentIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<string, string>>({});

  const { data: questions = [], isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentQuestions(templateId),
    queryFn: () => assessmentApi.getQuestionsByTemplate(templateId),
  });

  const submitMutation = useApiMutation({
    mutationFn: () => {
      const answerList: AnswerSubmit[] = Object.entries(answers).map(([questionId, optionId]) => ({
        questionId,
        optionId,
      }));
      return assessmentApi.submitAssessment({ templateId, userId: user!.id, answers: answerList });
    },
  });

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner className="h-6 w-6" />
      </div>
    );
  }

  if (submitMutation.data) {
    return (
      <div className="mx-auto max-w-2xl">
        <ResultSummary result={submitMutation.data} />
      </div>
    );
  }

  const sortedQuestions = [...questions].sort((a, b) => a.questionOrder - b.questionOrder);
  const currentQuestion = sortedQuestions[currentIndex];
  const isLastQuestion = currentIndex === sortedQuestions.length - 1;

  function handleAnswer(optionId: string) {
    setAnswers((current) => ({ ...current, [currentQuestion.slug]: optionId }));
  }

  return (
    <div className="mx-auto max-w-2xl">
      <div className="mb-6">
        <ProgressBar current={currentIndex} total={sortedQuestions.length} />
      </div>

      {currentQuestion && (
        <QuestionStep
          question={currentQuestion}
          selectedOptionId={answers[currentQuestion.slug]}
          onAnswer={handleAnswer}
        />
      )}

      <div className="mt-6 flex justify-between">
        <Button variant="outline" disabled={currentIndex === 0} onClick={() => setCurrentIndex((i) => i - 1)}>
          Back
        </Button>
        {isLastQuestion ? (
          <Button
            disabled={!currentQuestion || !answers[currentQuestion.slug] || submitMutation.isPending}
            onClick={() => submitMutation.mutate()}
          >
            {submitMutation.isPending ? "Submitting..." : "Submit"}
          </Button>
        ) : (
          <Button
            disabled={!currentQuestion || !answers[currentQuestion.slug]}
            onClick={() => setCurrentIndex((i) => i + 1)}
          >
            Next
          </Button>
        )}
      </div>
    </div>
  );
}
