import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";
import type {
  ExpertProfile,
  ExpertVerificationStatus,
  UpdateExpertProfileRequest,
  ServiceEnvelope,
  PaginatedResponse,
  Specialization,
} from "@/types";

/**
 * profile-service now returns a base Profile { id, slug, name, authId, role, ... }
 * with role-specific data nested under `expertProfile`. These helpers flatten
 * that shape back into the app's existing flat ExpertProfile type so callers
 * (find-experts, booking flow, admin/experts) don't need to change.
 */
interface RawExpertDetail {
  profileId: string;
  phoneNumber?: string;
  email?: string;
  avatarUrl?: string;
  introductionVideoUrl?: string;
  bio?: string;
  verificationStatus: ExpertVerificationStatus;
  specializations?: Specialization[];
}

interface RawExpertProfile {
  id: string;
  authId: string;
  name: string;
  expertProfile?: RawExpertDetail;
}

function toExpertProfile(raw: RawExpertProfile): ExpertProfile {
  const detail = raw.expertProfile;
  return {
    expertId: detail?.profileId ?? raw.id,
    accountId: raw.authId,
    fullName: raw.name,
    phoneNumber: detail?.phoneNumber,
    email: detail?.email ?? "",
    avatarUrl: detail?.avatarUrl || undefined,
    introductionVideoUrl: detail?.introductionVideoUrl,
    bio: detail?.bio,
    verificationStatus: detail?.verificationStatus ?? "UNVERIFIED",
    specializations: detail?.specializations ?? [],
  };
}

/** profile-service's UpsertExpertRequest expects `name`, not `fullName`. */
function toUpsertPayload(payload: UpdateExpertProfileRequest) {
  const { fullName, ...rest } = payload;
  return { name: fullName, ...rest };
}

export async function getMyProfile(): Promise<ExpertProfile> {
  const response = await profileClient.get<ServiceEnvelope<RawExpertProfile>>(PROFILE_ENDPOINTS.ME);
  return toExpertProfile(response.data.data);
}

export async function updateMyProfile(payload: UpdateExpertProfileRequest): Promise<ExpertProfile> {
  const response = await profileClient.put<ServiceEnvelope<RawExpertProfile>>(
    PROFILE_ENDPOINTS.ME,
    toUpsertPayload(payload)
  );
  return toExpertProfile(response.data.data);
}

export interface ListExpertsParams {
  page?: number;
  pageSize?: number;
  search?: string;
}

/** Real: GET /profiles/experts (public browsing, no server-side specialization filter). */
export async function listExperts(
  params: ListExpertsParams = {}
): Promise<PaginatedResponse<ExpertProfile>> {
  const response = await profileClient.get<ServiceEnvelope<PaginatedResponse<RawExpertProfile>>>(
    PROFILE_ENDPOINTS.EXPERTS,
    { params }
  );
  const page = response.data.data;
  return { ...page, items: page.items.map(toExpertProfile) };
}

/**
 * profile-service caps page_size at 100 server-side and has no true "get all"
 * route, so this fetches the largest single page — fine at current scale, see
 * listExperts for real pagination on the admin screen.
 */
export async function getAllExperts(): Promise<ExpertProfile[]> {
  const page = await listExperts({ pageSize: 100 });
  return page.items;
}

export interface ExpertSearchFilters {
  query?: string;
  specializationId?: string;
}

/**
 * profile-service has no server-side specialization filter yet, so this
 * fetches by name (server-side) and filters specialization client-side.
 */
export async function searchExperts(filters: ExpertSearchFilters): Promise<ExpertProfile[]> {
  const page = await listExperts({ pageSize: 100, search: filters.query });
  if (!filters.specializationId) return page.items;

  return page.items.filter((expert) =>
    expert.specializations.some((spec) => spec.specId === filters.specializationId)
  );
}

export async function getExpertProfile(accountId: string): Promise<ExpertProfile> {
  const response = await profileClient.get<ServiceEnvelope<RawExpertProfile>>(
    PROFILE_ENDPOINTS.EXPERT(accountId)
  );
  return toExpertProfile(response.data.data);
}

export async function getPublicProfile(accountId: string): Promise<ExpertProfile> {
  const response = await profileClient.get<ServiceEnvelope<RawExpertProfile>>(
    PROFILE_ENDPOINTS.PROFILE(accountId)
  );
  return toExpertProfile(response.data.data);
}

// ---------- Admin (PUT/PATCH /profiles/experts/{accountId}) ----------

export async function updateExpertProfile(
  accountId: string,
  payload: UpdateExpertProfileRequest
): Promise<ExpertProfile> {
  const response = await profileClient.put<ServiceEnvelope<RawExpertProfile>>(
    PROFILE_ENDPOINTS.EXPERT(accountId),
    toUpsertPayload(payload)
  );
  return toExpertProfile(response.data.data);
}

export async function updateExpertVerification(
  accountId: string,
  verificationStatus: ExpertVerificationStatus
): Promise<ExpertProfile> {
  const response = await profileClient.patch<ServiceEnvelope<RawExpertProfile>>(
    PROFILE_ENDPOINTS.EXPERT(accountId),
    { verificationStatus }
  );
  return toExpertProfile(response.data.data);
}
