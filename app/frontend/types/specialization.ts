export interface Specialization {
  specId: string;
  name: string;
  description?: string;
  imageUrl?: string;
  isActive: boolean;
}

export interface CreateSpecializationRequest {
  name: string;
  description?: string;
  imageUrl?: string;
}
