/** Documented in API-document.md; not backed by a real service yet — see /data/clinical-records.ts. */
export interface ClinicalRecord {
  recordId: string;
  appointmentId: string;
  clientId: string;
  diagnosis: string;
  notes: string;
  recommendations?: string;
  attachments: string[];
  createdAt: string;
}

export interface CreateClinicalRecordRequest {
  appointmentId: string;
  clientId: string;
  diagnosis: string;
  notes: string;
  recommendations?: string;
  attachments?: string[];
}
