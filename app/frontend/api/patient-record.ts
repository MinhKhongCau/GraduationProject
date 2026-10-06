import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";
import type { PatientRecord, CreatePatientRecordRequest } from "@/types";

/** [PATIENT] People the logged-in patient can book for; the SELF record always comes first. */
export async function getMyPatientRecords(): Promise<PatientRecord[]> {
  const response = await profileClient.get<PatientRecord[]>(PROFILE_ENDPOINTS.ME_PATIENT_RECORDS);
  return response.data ?? [];
}

/** [PATIENT] 409 when the account already has the maximum number of records. */
export async function createMyPatientRecord(payload: CreatePatientRecordRequest): Promise<PatientRecord> {
  const response = await profileClient.post<PatientRecord>(PROFILE_ENDPOINTS.ME_PATIENT_RECORDS, payload);
  return response.data;
}
