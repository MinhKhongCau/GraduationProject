export type AssessmentDimension = "DEPRESSION" | "ANXIETY" | "STRESS";

export interface AssessmentTemplate {
  code: string;
  title: string;
  description?: string;
  instruction?: string;
  certification?: string;
  slug: string;
  isActive?: boolean;
}

export interface AssessmentOption {
  label: string;
  value: string;
  scoreValue: number;
  orderIndex: number;
  slug: string;
}

export interface AssessmentQuestion {
  content: string;
  dimension: string;
  questionOrder: number;
  slug: string;
  options: AssessmentOption[];
}

export interface AnswerSubmit {
  questionId: string; // holds question slug
  optionId: string;   // holds option slug
}

export interface AssessmentSubmitRequest {
  templateId: string; // holds template slug
  userId: string;
  answers: AnswerSubmit[];
}

export interface AssessmentSubmitResponse {
  resultId: string;
  totalScore: number;
  dimensionScores: Record<string, number>;
  aiEvaluation: string;
}
