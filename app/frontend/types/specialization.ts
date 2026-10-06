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
