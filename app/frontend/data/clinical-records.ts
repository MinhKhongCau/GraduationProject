import type { ClinicalRecord } from "@/types";

/** No service implements clinical records yet — see DESIGN.md. */
export const CLINICAL_RECORDS_MOCK: ClinicalRecord[] = [
  {
    record_id: "record-1",
    recordId: "record-1",
    appointment_id: "appt-seed-1",
    appointmentId: "appt-seed-1",
    patient_id: "current-patient",
    patientId: "current-patient",
    expert_id: "expert-1",
    expertId: "expert-1",
    diagnosis: "Generalized Anxiety Disorder (GAD)",
    symptoms: "Persistent worry, fatigue, sleep disturbance",
    actions_to_avoid: "Caffeine in the evening, excessive screen time before sleep",
    actions_to_take: "Daily meditation 15 mins, breathing exercises",
    treatment_plan: "Cognitive Behavioral Therapy (CBT)",
    expert_notes: "Patient shows positive response to breathing exercises.",
    created_at: 1718874000000,
  },
];
