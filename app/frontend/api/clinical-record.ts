import { getMedicalRecords, saveMedicalRecord, getAppointmentMedicalRecord, getMedicalRecordDetail } from "./booking";
import type { MedicalRecord, SaveMedicalRecordRequest } from "@/types";

export async function getMyMedicalHistory(): Promise<MedicalRecord[]> {
  const result = await getMedicalRecords({ page: 0, size: 100 });
  return result.items;
}

export async function getClientMedicalHistory(clientId: string): Promise<MedicalRecord[]> {
  const result = await getMedicalRecords({ page: 0, size: 100 });
  return result.items.filter((record) => record.patient_id === clientId || record.patientId === clientId);
}

export async function createClinicalRecord(
  appointmentId: string,
  payload: SaveMedicalRecordRequest
): Promise<MedicalRecord> {
  return saveMedicalRecord(appointmentId, payload);
}

export {
  getMedicalRecords,
  saveMedicalRecord,
  getAppointmentMedicalRecord,
  getMedicalRecordDetail,
};

