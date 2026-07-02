import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";
import type {
  PatientProfile,
  UpdatePatientProfileRequest,
  MedicalHistory,
  CreateMedicalHistoryRequest,
  ServiceEnvelope,
} from "@/types";

export async function getPatientProfile(accountId: string): Promise<PatientProfile> {
  const response = await profileClient.get<ServiceEnvelope<PatientProfile>>(
    PROFILE_ENDPOINTS.PATIENT(accountId)
  );
  return response.data.data;
}

export async function updatePatientProfile(
  accountId: string,
  payload: UpdatePatientProfileRequest
): Promise<PatientProfile> {
  const response = await profileClient.post<ServiceEnvelope<PatientProfile>>(
    PROFILE_ENDPOINTS.PATIENT(accountId),
    payload
  );
  return response.data.data;
}

export async function getMedicalHistories(accountId: string): Promise<MedicalHistory[]> {
  const response = await profileClient.get<ServiceEnvelope<MedicalHistory[]>>(
    PROFILE_ENDPOINTS.PATIENT_MEDICAL_HISTORIES(accountId)
  );
  return response.data.data;
}

export async function addMedicalHistory(
  accountId: string,
  payload: CreateMedicalHistoryRequest
): Promise<MedicalHistory> {
  const response = await profileClient.post<ServiceEnvelope<MedicalHistory>>(
    PROFILE_ENDPOINTS.PATIENT_MEDICAL_HISTORIES(accountId),
    payload
  );
  return response.data.data;
}
