import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";
import { splitUserInformation, toPaginated, type RawUserInformation } from "./profile-shared";
import type {
  ExpertProfile,
  ExpertVerificationStatus,
  UpdateExpertProfileRequest,
  PaginatedResponse,
  PageResult,
  Specialization,
} from "@/types";

/**
 * profile-service returns a base Profile { id, slug, userInformation, authId, role, ... }
 * with role-specific data nested under `expertProfile`. These helpers flatten
 * that shape back into the app's existing flat ExpertProfile type so callers
 * (find-experts, booking flow, admin/experts) don't need to change.
 */
interface RawExpertDetail {
  profileId: string;
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
  userInformation: RawUserInformation;
  expertProfile?: RawExpertDetail;
}

function toExpertProfile(raw: RawExpertProfile): ExpertProfile {
  const detail = raw.expertProfile;
  const info = raw.userInformation;
  return {
    expertId: detail?.profileId ?? raw.id,
    accountId: raw.authId,
    fullName: info?.fullName ?? "",
    phoneNumber: info?.phoneNumber || undefined,
    dateOfBirth: info?.dateOfBirth ?? undefined,
    gender: info?.gender || undefined,
    country: info?.country || undefined,
    email: detail?.email ?? "",
    avatarUrl: detail?.avatarUrl || undefined,
    introductionVideoUrl: detail?.introductionVideoUrl,
    bio: detail?.bio,
    verificationStatus: detail?.verificationStatus ?? "UNVERIFIED",
    specializations: detail?.specializations ?? [],
  };
}

/** profile-service's UpsertExpertRequest nests personal info under `userInformation`. */
function toUpsertPayload(payload: UpdateExpertProfileRequest) {
  return splitUserInformation(payload);
}

export async function getMyProfile(): Promise<ExpertProfile> {
  const response = await profileClient.get<RawExpertProfile>(PROFILE_ENDPOINTS.ME);
  return toExpertProfile(response.data);
}

export async function updateMyProfile(payload: UpdateExpertProfileRequest): Promise<ExpertProfile> {
  const response = await profileClient.put<RawExpertProfile>(
    PROFILE_ENDPOINTS.ME,
    toUpsertPayload(payload)
  );
  return toExpertProfile(response.data);
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
  const response = await profileClient.get<PageResult<RawExpertProfile>>(
    PROFILE_ENDPOINTS.EXPERTS,
    { params }
  );
  return toPaginated(response.data, toExpertProfile);
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
  const response = await profileClient.get<RawExpertProfile>(
    PROFILE_ENDPOINTS.EXPERT(accountId)
  );
  return toExpertProfile(response.data);
}

export async function getPublicProfile(accountId: string): Promise<ExpertProfile> {
  const response = await profileClient.get<RawExpertProfile>(
    PROFILE_ENDPOINTS.PROFILE(accountId)
  );
  return toExpertProfile(response.data);
}

// ---------- Admin (PUT/PATCH /profiles/experts/{accountId}) ----------

export async function updateExpertProfile(
  accountId: string,
  payload: UpdateExpertProfileRequest
): Promise<ExpertProfile> {
  const response = await profileClient.put<RawExpertProfile>(
    PROFILE_ENDPOINTS.EXPERT(accountId),
    toUpsertPayload(payload)
  );
  return toExpertProfile(response.data);
}

export async function updateExpertVerification(
  accountId: string,
  verificationStatus: ExpertVerificationStatus
): Promise<ExpertProfile> {
  const response = await profileClient.patch<RawExpertProfile>(
    PROFILE_ENDPOINTS.EXPERT(accountId),
    { verificationStatus }
  );
  return toExpertProfile(response.data);
}
