export interface Specialization {
  specId: string;
  code: string;
  name: string;
  slug: string;
  description?: string;
  symptoms: string[];
  location?: string;
  imageUrl?: string;
  isActive: boolean;
}

export interface CreateSpecializationRequest {
  code: string;
  name: string;
  /** Generated from `name` by profile-service when omitted. */
  slug?: string;
  description?: string;
  symptoms?: string[];
  location?: string;
  imageUrl?: string;
}

/** Admin view of a specialization: includes inactive ones and how many experts use it. */
export interface AdminSpecialization extends Specialization {
  expertCount: number;
}

/** PUT replaces every field; a blank `slug` is regenerated from `name`. Status changes go through its own endpoint. */
export type UpdateSpecializationRequest = CreateSpecializationRequest;

export interface ListSpecializationsParams {
  page?: number;
  pageSize?: number;
  search?: string;
  isActive?: boolean;
}
