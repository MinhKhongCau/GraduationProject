import { assessmentClient } from "./http/instances";
import { ASSESSMENT_ENDPOINTS } from "@/constants/api";
import type {
  AssessmentTemplate,
  AssessmentQuestion,
  AssessmentDimension,
  AssessmentSubmitRequest,
  AssessmentSubmitResponse,
  AssessmentHistoryItem,
} from "@/types";

export async function getTemplates(): Promise<AssessmentTemplate[]> {
  const response = await assessmentClient.get<AssessmentTemplate[]>(ASSESSMENT_ENDPOINTS.TEMPLATES);
  return response.data;
}

export async function getTemplate(slug: string): Promise<AssessmentTemplate> {
  const response = await assessmentClient.get<AssessmentTemplate>(
    ASSESSMENT_ENDPOINTS.TEMPLATE(slug)
  );
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

export async function getMyAssessments(): Promise<AssessmentHistoryItem[]> {
  const response = await assessmentClient.get<AssessmentHistoryItem[]>(ASSESSMENT_ENDPOINTS.SELF);
  return response.data;
}

// Admin Templates Management
export async function createTemplate(payload: {
  code: string;
  title: string;
  description?: string;
  instruction?: string;
  certification?: string;
}): Promise<any> {
  const response = await assessmentClient.post(ASSESSMENT_ENDPOINTS.TEMPLATES, payload);
  return response.data;
}

export async function updateTemplate(
  slug: string,
  payload: {
    code?: string;
    title?: string;
    description?: string;
    instruction?: string;
    certification?: string;
    isActive?: boolean;
  }
): Promise<any> {
  const response = await assessmentClient.patch(ASSESSMENT_ENDPOINTS.TEMPLATE(slug), payload);
  return response.data;
}

export async function deleteTemplate(slug: string): Promise<any> {
  const response = await assessmentClient.delete(ASSESSMENT_ENDPOINTS.TEMPLATE(slug));
  return response.data;
}

// Admin Dimensions Management
export async function getDimensions(): Promise<AssessmentDimension[]> {
  const response = await assessmentClient.get<AssessmentDimension[]>(ASSESSMENT_ENDPOINTS.DIMENSIONS);
  return response.data;
}

export async function getDimension(slug: string): Promise<AssessmentDimension> {
  const response = await assessmentClient.get<AssessmentDimension>(
    ASSESSMENT_ENDPOINTS.DIMENSION(slug)
  );
  return response.data;
}

export async function createDimension(payload: {
  code: string;
  name: string;
  description?: string;
}): Promise<any> {
  const response = await assessmentClient.post(ASSESSMENT_ENDPOINTS.DIMENSIONS, payload);
  return response.data;
}

export async function updateDimension(
  slug: string,
  payload: { code?: string; name?: string; description?: string; isActive?: boolean }
): Promise<any> {
  const response = await assessmentClient.patch(ASSESSMENT_ENDPOINTS.DIMENSION(slug), payload);
  return response.data;
}

export async function deleteDimension(slug: string): Promise<any> {
  const response = await assessmentClient.delete(ASSESSMENT_ENDPOINTS.DIMENSION(slug));
  return response.data;
}

export async function bulkAssignQuestionsToDimension(
  slug: string,
  questionIds: string[] // question slugs
): Promise<any> {
  const response = await assessmentClient.patch(ASSESSMENT_ENDPOINTS.DIMENSION_QUESTIONS_BULK(slug), {
    questionIds,
  });
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
  questions: { content: string; dimensionId: string; questionOrder: number }[]; // dimensionId holds dimension slug
}): Promise<any> {
  const response = await assessmentClient.post(ASSESSMENT_ENDPOINTS.BULK_QUESTIONS, payload);
  return response.data;
}

export async function updateQuestion(
  slug: string,
  payload: { content?: string; dimensionId?: string; questionOrder?: number; isRequired?: boolean }
): Promise<any> {
  const response = await assessmentClient.patch(ASSESSMENT_ENDPOINTS.QUESTION(slug), payload);
  return response.data;
}

export async function deleteQuestion(slug: string): Promise<any> {
  const response = await assessmentClient.delete(ASSESSMENT_ENDPOINTS.QUESTION(slug));
  return response.data;
}
