import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";
import { toPaginated } from "./profile-shared";
import type {
  Specialization,
  CreateSpecializationRequest,
  AdminSpecialization,
  UpdateSpecializationRequest,
  ListSpecializationsParams,
  PageResult,
  PaginatedResponse,
} from "@/types";

/** [PUBLIC] Active specializations only. */
export async function getAllSpecializations(): Promise<Specialization[]> {
  const response = await profileClient.get<Specialization[]>(
    PROFILE_ENDPOINTS.SPECIALIZATIONS
  );
  return response.data ?? [];
}

export async function createSpecialization(
  payload: CreateSpecializationRequest
): Promise<Specialization> {
  const response = await profileClient.post<Specialization>(
    PROFILE_ENDPOINTS.SPECIALIZATIONS,
    payload
  );
  return response.data;
}

// ---------- Admin ----------

/** [ADMIN] Every specialization, including inactive ones, with the number of experts using each. */
export async function listAllSpecializations(
  params: ListSpecializationsParams = {}
): Promise<PaginatedResponse<AdminSpecialization>> {
  const response = await profileClient.get<PageResult<AdminSpecialization>>(
    PROFILE_ENDPOINTS.SPECIALIZATIONS_ALL,
    { params }
  );
  return toPaginated(response.data, (spec) => spec);
}

export async function getSpecialization(specId: string): Promise<AdminSpecialization> {
  const response = await profileClient.get<AdminSpecialization>(
    PROFILE_ENDPOINTS.SPECIALIZATION(specId)
  );
  return response.data;
}

/** [ADMIN] 409 when the code or slug is taken by another specialization. */
export async function updateSpecialization(
  specId: string,
  payload: UpdateSpecializationRequest
): Promise<AdminSpecialization> {
  const response = await profileClient.put<AdminSpecialization>(
    PROFILE_ENDPOINTS.SPECIALIZATION(specId),
    payload
  );
  return response.data;
}

/** [ADMIN] Inactive specializations are hidden from patients; experts already using one keep it. */
export async function updateSpecializationStatus(
  specId: string,
  isActive: boolean
): Promise<AdminSpecialization> {
  const response = await profileClient.patch<AdminSpecialization>(
    PROFILE_ENDPOINTS.SPECIALIZATION_STATUS(specId),
    { isActive }
  );
  return response.data;
}

/** [ADMIN] 409 while any expert still uses it — deactivate instead. */
export async function deleteSpecialization(specId: string): Promise<void> {
  await profileClient.delete(PROFILE_ENDPOINTS.SPECIALIZATION(specId));
}
