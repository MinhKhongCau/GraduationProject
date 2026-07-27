import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";
import type {
  PatientProfile,
  UpdatePatientProfileRequest,
  MedicalHistory,
  CreateMedicalHistoryRequest,
  ServiceEnvelope,
  PaginatedResponse,
} from "@/types";

/**
 * profile-service now returns a base Profile { id, slug, name, authId, role, ... }
 * with role-specific data nested under `patientProfile`. These helpers flatten
 * that shape back into the app's existing flat PatientProfile type so callers
 * (settings, medical-history, admin/patients) don't need to change.
 */
interface RawPatientDetail {
  profileId: string;
  phoneNumber?: string;
  email?: string;
  avatarUrl?: string;
  dateOfBirth?: string | null;
  gender?: string;
  address?: string;
  medicalHistories?: MedicalHistory[];
}

interface RawPatientProfile {
  id: string;
  authId: string;
  name: string;
  patientProfile?: RawPatientDetail;
}

function toPatientProfile(raw: RawPatientProfile): PatientProfile {
  const detail = raw.patientProfile;
  return {
    patientId: detail?.profileId ?? raw.id,
    accountId: raw.authId,
    fullName: raw.name,
    phoneNumber: detail?.phoneNumber,
    email: detail?.email ?? "",
    avatarUrl: detail?.avatarUrl || undefined,
    dateOfBirth: detail?.dateOfBirth ?? undefined,
    gender: detail?.gender,
    address: detail?.address,
    medicalHistories: detail?.medicalHistories,
  };
}

/** profile-service's UpsertPatientRequest expects `name`, not `fullName`. */
function toUpsertPayload(payload: UpdatePatientProfileRequest) {
  const { fullName, ...rest } = payload;
  return { name: fullName, ...rest };
}

// ---------- Self-service (GET/PUT /profiles/me) ----------

export async function getMyProfile(): Promise<PatientProfile> {
  const response = await profileClient.get<ServiceEnvelope<RawPatientProfile>>(PROFILE_ENDPOINTS.ME);
  return toPatientProfile(response.data.data);
}

export async function updateMyProfile(payload: UpdatePatientProfileRequest): Promise<PatientProfile> {
  const response = await profileClient.put<ServiceEnvelope<RawPatientProfile>>(
    PROFILE_ENDPOINTS.ME,
    toUpsertPayload(payload)
  );
  return toPatientProfile(response.data.data);
}

export async function getMedicalHistories(): Promise<MedicalHistory[]> {
  const response = await profileClient.get<ServiceEnvelope<MedicalHistory[]>>(
    PROFILE_ENDPOINTS.ME_MEDICAL_HISTORIES
  );
  return response.data.data ?? [];
}

export async function addMedicalHistory(
  payload: CreateMedicalHistoryRequest
): Promise<MedicalHistory> {
  const response = await profileClient.post<ServiceEnvelope<MedicalHistory>>(
    PROFILE_ENDPOINTS.ME_MEDICAL_HISTORIES,
    payload
  );
  return response.data.data;
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
  const response = await profileClient.get<ServiceEnvelope<PaginatedResponse<RawPatientProfile>>>(
    PROFILE_ENDPOINTS.PATIENTS,
    { params }
  );
  const page = response.data.data;
  return { ...page, items: page.items.map(toPatientProfile) };
}

export async function getPatientProfile(accountId: string): Promise<PatientProfile> {
  const response = await profileClient.get<ServiceEnvelope<RawPatientProfile>>(
    PROFILE_ENDPOINTS.PATIENT(accountId)
  );
  return toPatientProfile(response.data.data);
}

export async function getPatientProfilePublic(accountId: string): Promise<PatientProfile> {
  const response = await profileClient.get<ServiceEnvelope<RawPatientProfile>>(
    PROFILE_ENDPOINTS.PATIENT_PUBLIC(accountId)
  );
  return toPatientProfile(response.data.data);
}

export async function updatePatientProfile(
  accountId: string,
  payload: UpdatePatientProfileRequest
): Promise<PatientProfile> {
  const response = await profileClient.put<ServiceEnvelope<RawPatientProfile>>(
    PROFILE_ENDPOINTS.PATIENT(accountId),
    toUpsertPayload(payload)
  );
  return toPatientProfile(response.data.data);
}
