import { CLINICAL_RECORDS_MOCK } from "@/data/clinical-records";
import type { ClinicalRecord, CreateClinicalRecordRequest } from "@/types";

/**
 * Documented in API-document.md; no backend service implements clinical
 * records yet (see DESIGN.md), so these always use the mock dataset.
 */

export async function getMyMedicalHistory(): Promise<ClinicalRecord[]> {
  return Promise.resolve(CLINICAL_RECORDS_MOCK);
}

export async function getClientMedicalHistory(clientId: string): Promise<ClinicalRecord[]> {
  return Promise.resolve(CLINICAL_RECORDS_MOCK.filter((record) => record.clientId === clientId));
}

export async function createClinicalRecord(
  payload: CreateClinicalRecordRequest
): Promise<ClinicalRecord> {
  const record: ClinicalRecord = {
    recordId: `record-${Date.now()}`,
    ...payload,
    attachments: payload.attachments ?? [],
    createdAt: new Date().toISOString(),
  };
  CLINICAL_RECORDS_MOCK.push(record);
  return Promise.resolve(record);
}
