export interface MedicalRecord {
  record_id: string;
  recordId?: string;
  appointment_id: string;
  appointmentId?: string;
  patient_id: string;
  patientId?: string;
  expert_id: string;
  expertId?: string;
  diagnosis: string;             // Tình trạng bệnh / Chẩn đoán
  symptoms: string;              // Triệu chứng
  actions_to_avoid: string;      // Các hành động cần tránh
  actionsToAvoid?: string;
  actions_to_take: string;       // Các hành động cần làm
  actionsToTake?: string;
  treatment_plan?: string;       // Phác đồ / Kế hoạch điều trị
  treatmentPlan?: string;
  next_appointment_date?: number | null; // Unix timestamp ms
  nextAppointmentDate?: number | null;
  next_appointment_note?: string;
  nextAppointmentNote?: string;
  expert_notes?: string;
  expertNotes?: string;
  created_at: number;
  createdAt?: number | string;
  updated_at?: number;
  updatedAt?: number | string;

  // Extra joined fields
  appointment_date?: number;
  appointmentDate?: number;
  appointment_price?: number;
  meeting_link?: string;
  appointment_status?: string;

  // Populated UI display fields
  patientName?: string;
  patientEmail?: string;
  patientAvatar?: string;
  expertName?: string;
  expertAvatar?: string;
}

export interface SaveMedicalRecordRequest {
  diagnosis: string;
  symptoms: string;
  actions_to_avoid: string;
  actions_to_take: string;
  treatment_plan?: string;
  next_appointment_date?: number | null;
  next_appointment_note?: string;
  expert_notes?: string;
}

// Backward-compatibility alias
export type ClinicalRecord = MedicalRecord;
export type CreateClinicalRecordRequest = SaveMedicalRecordRequest;
