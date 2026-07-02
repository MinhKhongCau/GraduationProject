import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";
import type { Specialization, CreateSpecializationRequest, ServiceEnvelope } from "@/types";

export async function getAllSpecializations(): Promise<Specialization[]> {
  const response = await profileClient.get<ServiceEnvelope<Specialization[]>>(
    PROFILE_ENDPOINTS.SPECIALIZATIONS
  );
  return response.data.data ?? [];
}

export async function createSpecialization(
  payload: CreateSpecializationRequest
): Promise<Specialization> {
  const response = await profileClient.post<ServiceEnvelope<Specialization>>(
    PROFILE_ENDPOINTS.SPECIALIZATIONS,
    payload
  );
  return response.data.data;
}
