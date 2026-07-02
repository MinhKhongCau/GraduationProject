export type AssessmentDimension = "DEPRESSION" | "ANXIETY" | "STRESS";

export interface AssessmentTemplate {
  templateId: string;
  code: string;
  title: string;
  description?: string;
}

export interface AssessmentOption {
  optionId: string;
  label: string;
  scoreValue: number;
  orderIndex: number;
}

export interface AssessmentQuestion {
  questionId: string;
  content: string;
  dimension: AssessmentDimension;
  questionOrder: number;
  options: AssessmentOption[];
}

export interface AnswerSubmit {
  questionId: string;
  optionId: string;
}

export interface AssessmentSubmitRequest {
  templateId: string;
  userId: string;
  answers: AnswerSubmit[];
}

export interface AssessmentSubmitResponse {
  resultId: string;
  totalScore: number;
  dimensionScores: Record<string, number>;
  aiEvaluation: string;
}
