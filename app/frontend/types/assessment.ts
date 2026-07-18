export interface AssessmentDimension {
  code: string;
  name: string;
  description?: string;
  slug: string;
  isActive?: boolean;
}

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
  questionContent: string; // question text, used to build the AI prompt
  optionLabel: string;     // selected answer label, used to build the AI prompt
}

export interface AssessmentSubmitRequest {
  templateId: string; // holds template slug
  answers: AnswerSubmit[];
}

export interface AssessmentSubmitResponse {
  resultId: string;
  totalScore: number;
  dimensionScores: Record<string, number>;
  aiEvaluation: string;
}

export interface AssessmentHistoryItem {
  resultId: string;
  templateCode: string;
  templateTitle: string;
  totalScore: number;
  dimensionScores: Record<string, number>;
  aiEvaluation?: string;
  createdAt: string;
}
