import { assessmentClient } from "./http/instances";
import { ASSESSMENT_ENDPOINTS } from "@/constants/api";
import type { AssessmentTemplate, AssessmentQuestion, AssessmentSubmitRequest, AssessmentSubmitResponse } from "@/types";

export async function getTemplates(): Promise<AssessmentTemplate[]> {
  const response = await assessmentClient.get<AssessmentTemplate[]>(ASSESSMENT_ENDPOINTS.TEMPLATES);
  return response.data;
}

export async function getQuestionsByTemplate(templateId: string): Promise<AssessmentQuestion[]> {
  const response = await assessmentClient.get<AssessmentQuestion[]>(
    ASSESSMENT_ENDPOINTS.QUESTIONS(templateId)
  );
  return response.data;
}

export async function submitAssessment(
  payload: AssessmentSubmitRequest
): Promise<AssessmentSubmitResponse> {
  // Authorization header is attached automatically by the request
  // interceptor; assessment-service's stub auth just checks for its
  // presence (see DESIGN.md).
  const response = await assessmentClient.post<AssessmentSubmitResponse>(
    ASSESSMENT_ENDPOINTS.SUBMIT,
    payload
  );
  return response.data;
}
