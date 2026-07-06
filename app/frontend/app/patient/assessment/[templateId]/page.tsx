"use client";

import { use, useState } from "react";
import { QuestionStep } from "./component/QuestionStep";
import { ProgressBar } from "./component/ProgressBar";
import { ResultSummary } from "./component/ResultSummary";
import { Button, Spinner, Card } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { assessmentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";
import { BookOpen, Award, ChevronDown, ChevronUp, FileText, Play } from "lucide-react";
import type { AnswerSubmit } from "@/types";

// Safe, lightweight custom markdown parser
function MarkdownRenderer({ text }: { text: string }) {
  if (!text) return null;
  
  const lines = text.split("\n");
  const processed = lines.map((line, idx) => {
    const trimmed = line.trim();
    if (trimmed.startsWith("# ")) {
      return <h1 key={idx} className="text-lg font-bold text-foreground mt-4 mb-2">{trimmed.slice(2)}</h1>;
    }
    if (trimmed.startsWith("## ")) {
      return <h2 key={idx} className="text-base font-bold text-foreground mt-3 mb-2">{trimmed.slice(3)}</h2>;
    }
    if (trimmed.startsWith("### ")) {
      return <h3 key={idx} className="text-sm font-bold text-foreground mt-2 mb-1">{trimmed.slice(4)}</h3>;
    }
    if (trimmed.startsWith("- ") || trimmed.startsWith("* ")) {
      return <li key={idx} className="ml-4 list-disc text-sm text-muted-foreground mb-1.5">{trimmed.slice(2)}</li>;
    }
    if (trimmed === "") {
      return <div key={idx} className="h-2" />;
    }
    
    // Bold/Italic parser
    const formattedHtml = trimmed
      .replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>")
      .replace(/\*(.*?)\*/g, "<em>$1</em>");
      
    return (
      <p 
        key={idx} 
        className="text-sm text-muted-foreground leading-relaxed mb-2"
        dangerouslySetInnerHTML={{ __html: formattedHtml }}
      />
    );
  });

  return <div className="space-y-1">{processed}</div>;
}

export default function AssessmentDetailPage({ params }: { params: Promise<{ templateId: string }> }) {
  const { templateId } = use(params);
  const { user } = useAuthContext();

  const [isStarted, setIsStarted] = useState(false);
  const [showInfo, setShowInfo] = useState(false);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<string, string>>({});

  // Fetch Template Detail
  const { data: template, isLoading: isTemplateLoading } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentTemplate(templateId),
    queryFn: () => assessmentApi.getTemplate(templateId),
  });

  // Fetch Questions
  const { data: questions = [], isLoading: isQuestionsLoading } = useApiQuery({
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

  const isLoading = isTemplateLoading || isQuestionsLoading;

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner className="h-6 w-6" />
      </div>
    );
  }

  if (!template) {
    return (
      <Card className="mx-auto max-w-2xl p-6 text-center text-muted-foreground">
        Không tìm thấy thông tin bài đánh giá.
      </Card>
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

  // Welcome / Intro view
  if (!isStarted) {
    return (
      <div className="mx-auto max-w-2xl space-y-6">
        <Card className="p-6 md:p-8 space-y-6">
          <div className="border-b border-border/60 pb-5">
            <span className="inline-flex rounded-full bg-primary-soft px-3 py-1 text-xs font-semibold text-primary uppercase mb-3">
              {template.code}
            </span>
            <h1 className="text-2xl font-bold text-foreground">{template.title}</h1>
            <p className="mt-2.5 text-sm text-muted-foreground leading-relaxed">
              {template.description || "Bài đánh giá tâm lý lâm sàng tiêu chuẩn."}
            </p>
          </div>

          {/* Instruction Block */}
          {template.instruction && (
            <div className="space-y-3">
              <h3 className="flex items-center gap-2 text-sm font-bold text-foreground">
                <BookOpen className="h-4.5 w-4.5 text-primary" /> Hướng dẫn làm bài
              </h3>
              <div className="rounded-xl bg-surface/50 border border-border p-4.5">
                <MarkdownRenderer text={template.instruction} />
              </div>
            </div>
          )}

          {/* Certification Block */}
          {template.certification && (
            <div className="space-y-3">
              <h3 className="flex items-center gap-2 text-sm font-bold text-foreground">
                <Award className="h-4.5 w-4.5 text-success" /> Chứng nhận chuyên môn & Cơ sở khoa học
              </h3>
              <div className="rounded-xl bg-success-soft/20 border border-success/10 p-4.5">
                <MarkdownRenderer text={template.certification} />
              </div>
            </div>
          )}

          <div className="pt-4 flex justify-end">
            <Button 
              size="lg" 
              onClick={() => setIsStarted(true)} 
              className="w-full sm:w-auto flex items-center justify-center gap-2 font-semibold"
            >
              <Play className="h-4 w-4 fill-current" /> Bắt đầu đánh giá
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-4">
      {/* Collapsible reference drawer during test */}
      <Card className="overflow-hidden border border-border/80">
        <button 
          onClick={() => setShowInfo(!showInfo)} 
          className="flex w-full items-center justify-between bg-surface/40 px-5 py-3 hover:bg-surface/75 transition-colors"
        >
          <div className="flex items-center gap-2 text-xs font-semibold text-muted-foreground">
            <FileText className="h-4 w-4 text-primary" />
            <span>Thông tin & Hướng dẫn: {template.title}</span>
          </div>
          {showInfo ? <ChevronUp className="h-4 w-4 text-muted-foreground" /> : <ChevronDown className="h-4 w-4 text-muted-foreground" />}
        </button>

        {showInfo && (
          <div className="border-t border-border bg-background p-5 space-y-4 max-h-[300px] overflow-y-auto">
            {template.description && (
              <p className="text-xs text-muted-foreground italic">{template.description}</p>
            )}
            {template.instruction && (
              <div className="space-y-1.5">
                <h4 className="text-xs font-bold text-foreground">Hướng dẫn:</h4>
                <div className="rounded-lg bg-surface/30 p-3 border border-border/40">
                  <MarkdownRenderer text={template.instruction} />
                </div>
              </div>
            )}
            {template.certification && (
              <div className="space-y-1.5">
                <h4 className="text-xs font-bold text-foreground">Chứng nhận & Cơ sở khoa học:</h4>
                <div className="rounded-lg bg-success-soft/10 p-3 border border-success/5">
                  <MarkdownRenderer text={template.certification} />
                </div>
              </div>
            )}
          </div>
        )}
      </Card>

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
          Quay lại
        </Button>
        {isLastQuestion ? (
          <Button
            disabled={!currentQuestion || !answers[currentQuestion.slug] || submitMutation.isPending}
            onClick={() => submitMutation.mutate()}
          >
            {submitMutation.isPending ? "Đang gửi..." : "Hoàn thành"}
          </Button>
        ) : (
          <Button
            disabled={!currentQuestion || !answers[currentQuestion.slug]}
            onClick={() => setCurrentIndex((i) => i + 1)}
          >
            Tiếp theo
          </Button>
        )}
      </div>
    </div>
  );
}
