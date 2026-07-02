import type { ClinicalRecord } from "@/types";

/** No service implements clinical records yet — see DESIGN.md. */
export const CLINICAL_RECORDS_MOCK: ClinicalRecord[] = [
  {
    recordId: "record-1",
    appointmentId: "appt-seed-1",
    clientId: "current-patient",
    diagnosis: "Generalized Anxiety Disorder (GAD)",
    notes: "Patient shows persistent anxiety symptoms and requires further monitoring.",
    recommendations: "Practice daily meditation and schedule a follow-up session in two weeks.",
    attachments: [],
    createdAt: new Date(2026, 5, 20, 10, 0).toISOString(),
  },
];
