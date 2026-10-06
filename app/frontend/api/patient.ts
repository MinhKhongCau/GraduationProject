import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";
import { splitUserInformation, toPaginated, type RawUserInformation } from "./profile-shared";
import type {
  PatientProfile,
  UpdatePatientProfileRequest,
  MedicalHistory,
  CreateMedicalHistoryRequest,
  PaginatedResponse,
  PageResult,
} from "@/types";

/**
 * profile-service returns a base Profile { id, slug, userInformation, authId, role, ... }
 * with role-specific data nested under `patientProfile`. These helpers flatten
 * that shape back into the app's existing flat PatientProfile type so callers
 * (settings, medical-history, admin/patients) don't need to change.
 */
interface RawPatientDetail {
  profileId: string;
  email?: string;
  avatarUrl?: string;
  address?: string;
  medicalHistories?: MedicalHistory[];
}

interface RawPatientProfile {
  id: string;
  authId: string;
  userInformation: RawUserInformation;
  patientProfile?: RawPatientDetail;
}

function toPatientProfile(raw: RawPatientProfile): PatientProfile {
  const detail = raw.patientProfile;
  const info = raw.userInformation;
  return {
    patientId: detail?.profileId ?? raw.id,
    accountId: raw.authId,
    fullName: info?.fullName ?? "",
    phoneNumber: info?.phoneNumber || undefined,
    email: detail?.email ?? "",
    avatarUrl: detail?.avatarUrl || undefined,
    dateOfBirth: info?.dateOfBirth ?? undefined,
    gender: info?.gender || undefined,
    country: info?.country || undefined,
    address: detail?.address,
    medicalHistories: detail?.medicalHistories,
  };
}

/** profile-service's UpsertPatientRequest nests personal info under `userInformation`. */
function toUpsertPayload(payload: UpdatePatientProfileRequest) {
  return splitUserInformation(payload);
}

// ---------- Self-service (GET/PUT /profiles/me) ----------

export async function getMyProfile(): Promise<PatientProfile> {
  const response = await profileClient.get<RawPatientProfile>(PROFILE_ENDPOINTS.ME);
  return toPatientProfile(response.data);
}

export async function updateMyProfile(payload: UpdatePatientProfileRequest): Promise<PatientProfile> {
  const response = await profileClient.put<RawPatientProfile>(
    PROFILE_ENDPOINTS.ME,
    toUpsertPayload(payload)
  );
  return toPatientProfile(response.data);
}

export async function getMedicalHistories(): Promise<MedicalHistory[]> {
  const response = await profileClient.get<MedicalHistory[]>(
    PROFILE_ENDPOINTS.ME_MEDICAL_HISTORIES
  );
  return response.data ?? [];
}

export async function addMedicalHistory(
  payload: CreateMedicalHistoryRequest
): Promise<MedicalHistory> {
  const response = await profileClient.post<MedicalHistory>(
    PROFILE_ENDPOINTS.ME_MEDICAL_HISTORIES,
    payload
  );
  return response.data;
}

// ---------- Admin (GET/PUT /profiles/patients) ----------

export interface ListPatientsParams {
  page?: number;
  pageSize?: number;
  search?: string;
}

export async function listPatients(
  params: ListPatientsParams = {}
): Promise<PaginatedResponse<PatientProfile>> {
  const response = await profileClient.get<PageResult<RawPatientProfile>>(
    PROFILE_ENDPOINTS.PATIENTS,
    { params }
  );
  return toPaginated(response.data, toPatientProfile);
}

export async function getPatientProfile(accountId: string): Promise<PatientProfile> {
  const response = await profileClient.get<RawPatientProfile>(
    PROFILE_ENDPOINTS.PATIENT(accountId)
  );
  return toPatientProfile(response.data);
}

export async function getPatientProfilePublic(accountId: string): Promise<PatientProfile> {
  const response = await profileClient.get<RawPatientProfile>(
    PROFILE_ENDPOINTS.PATIENT_PUBLIC(accountId)
  );
  return toPatientProfile(response.data);
}

export async function updatePatientProfile(
  accountId: string,
  payload: UpdatePatientProfileRequest
): Promise<PatientProfile> {
  const response = await profileClient.put<RawPatientProfile>(
    PROFILE_ENDPOINTS.PATIENT(accountId),
    toUpsertPayload(payload)
  );
  return toPatientProfile(response.data);
}
