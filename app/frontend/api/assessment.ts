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
  const response = await assessmentClient.post<AssessmentSubmitResponse>(
    ASSESSMENT_ENDPOINTS.SUBMIT,
    payload
  );
  return response.data;
}

// Admin Templates Management
export async function createTemplate(payload: {
  code: string;
  title: string;
  description?: string;
}): Promise<any> {
  const response = await assessmentClient.post(ASSESSMENT_ENDPOINTS.TEMPLATES, payload);
  return response.data;
}

export async function updateTemplate(
  slug: string,
  payload: { code?: string; title?: string; description?: string; isActive?: boolean }
): Promise<any> {
  const response = await assessmentClient.patch(ASSESSMENT_ENDPOINTS.TEMPLATE(slug), payload);
  return response.data;
}

export async function deleteTemplate(slug: string): Promise<any> {
  const response = await assessmentClient.delete(ASSESSMENT_ENDPOINTS.TEMPLATE(slug));
  return response.data;
}

// Admin Option Groups Management
export async function getOptionGroups(): Promise<any[]> {
  const response = await assessmentClient.get<any[]>(ASSESSMENT_ENDPOINTS.OPTION_GROUPS);
  return response.data;
}

export async function getOptionGroup(slug: string): Promise<any> {
  const response = await assessmentClient.get<any>(ASSESSMENT_ENDPOINTS.OPTION_GROUP(slug));
  return response.data;
}

export async function createOptionGroup(payload: {
  groupCode: string;
  groupName: string;
  description?: string;
  options: { label: string; value: string; scoreValue: number; orderIndex: number }[];
}): Promise<any> {
  const response = await assessmentClient.post(ASSESSMENT_ENDPOINTS.OPTION_GROUPS, payload);
  return response.data;
}

export async function updateOptionGroup(
  slug: string,
  payload: { groupCode?: string; groupName?: string; description?: string }
): Promise<any> {
  const response = await assessmentClient.patch(ASSESSMENT_ENDPOINTS.OPTION_GROUP(slug), payload);
  return response.data;
}

export async function deleteOptionGroup(slug: string): Promise<any> {
  const response = await assessmentClient.delete(ASSESSMENT_ENDPOINTS.OPTION_GROUP(slug));
  return response.data;
}

// Admin Questions Management
export async function getQuestion(slug: string): Promise<any> {
  const response = await assessmentClient.get<any>(ASSESSMENT_ENDPOINTS.QUESTION(slug));
  return response.data;
}

export async function createBulkQuestions(payload: {
  templateId: string; // template slug
  groupId: string; // option group slug
  questions: { content: string; dimension: string; questionOrder: number }[];
}): Promise<any> {
  const response = await assessmentClient.post(ASSESSMENT_ENDPOINTS.BULK_QUESTIONS, payload);
  return response.data;
}

export async function updateQuestion(
  slug: string,
  payload: { content?: string; dimension?: string; questionOrder?: number; isRequired?: boolean }
): Promise<any> {
  const response = await assessmentClient.patch(ASSESSMENT_ENDPOINTS.QUESTION(slug), payload);
  return response.data;
}

export async function deleteQuestion(slug: string): Promise<any> {
  const response = await assessmentClient.delete(ASSESSMENT_ENDPOINTS.QUESTION(slug));
  return response.data;
}
