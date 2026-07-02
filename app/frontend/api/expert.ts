import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";
import { getMockWeeklySchedule, setMockWeeklySchedule } from "@/data/schedule";
import type {
  ExpertProfile,
  CreateExpertProfileRequest,
  ServiceEnvelope,
  WeeklyScheduleRequest,
  WeeklySlotInput,
} from "@/types";

/** Real: GET /profiles/experts/ (no server-side filters yet — see searchExperts). */
export async function getAllExperts(): Promise<ExpertProfile[]> {
  const response = await profileClient.get<ServiceEnvelope<ExpertProfile[]>>(
    PROFILE_ENDPOINTS.EXPERTS
  );
  return response.data.data ?? [];
}

export interface ExpertSearchFilters {
  query?: string;
  specializationId?: string;
}

/**
 * profile-service has no server-side search/filter support yet, so this
 * fetches the full list and filters client-side.
 */
export async function searchExperts(filters: ExpertSearchFilters): Promise<ExpertProfile[]> {
  const experts = await getAllExperts();
  const query = filters.query?.trim().toLowerCase();

  return experts.filter((expert) => {
    const matchesQuery =
      !query ||
      expert.fullName.toLowerCase().includes(query) ||
      expert.specializations.some((spec) => spec.name.toLowerCase().includes(query));
    const matchesSpecialization =
      !filters.specializationId ||
      expert.specializations.some((spec) => spec.specId === filters.specializationId);
    return matchesQuery && matchesSpecialization;
  });
}

export async function getExpertProfile(accountId: string): Promise<ExpertProfile> {
  const response = await profileClient.get<ServiceEnvelope<ExpertProfile>>(
    PROFILE_ENDPOINTS.EXPERT(accountId)
  );
  return response.data.data;
}

export async function createExpertProfile(
  accountId: string,
  payload: CreateExpertProfileRequest
): Promise<ExpertProfile> {
  const response = await profileClient.post<ServiceEnvelope<ExpertProfile>>(
    PROFILE_ENDPOINTS.EXPERT(accountId),
    payload
  );
  return response.data.data;
}

/**
 * Documented in API-document.md as POST /experts/me/schedule; booking-service
 * has no handler yet, so this reads/writes the in-memory mock instead of
 * hitting bookingClient.
 */
export async function getWeeklySchedule(): Promise<WeeklySlotInput[]> {
  return Promise.resolve(getMockWeeklySchedule());
}

export async function updateWeeklySchedule(
  payload: WeeklyScheduleRequest
): Promise<WeeklySlotInput[]> {
  setMockWeeklySchedule(payload.weeklySlots);
  return Promise.resolve(payload.weeklySlots);
}
