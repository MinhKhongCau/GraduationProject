import type { Appointment } from "@/types";

/**
 * booking-service has no appointment create/lock/history endpoints yet
 * (only slot generation/query — see DESIGN.md). This in-memory store lets
 * the book-appointment wizard and my-bookings page feel real within a
 * session. Deliberately NOT persisted across a full page reload.
 */
let mockAppointments: Appointment[] = [
  {
    appointmentId: "appt-seed-1",
    slotId: "slot-seed-1",
    patientId: "current-patient",
    expertId: "featured-2",
    expertName: "Dr. Tran Minh Tam",
    topic: "Anxiety & Panic",
    status: "CONFIRMED",
    createdAt: new Date(2026, 5, 20, 9, 0).toISOString(),
  },
];

export function listMockAppointments(): Appointment[] {
  return [...mockAppointments].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

export function getNextQueueNumber(slotId: string): number {
  return mockAppointments.filter((appointment) => appointment.slotId === slotId).length + 1;
}

export interface AddMockAppointmentInput {
  slotId: string;
  expertId: string;
  expertName: string;
  topic: string;
}

export function addMockAppointment(input: AddMockAppointmentInput): Appointment {
  const appointment: Appointment = {
    appointmentId: `appt-${Date.now()}`,
    slotId: input.slotId,
    patientId: "current-patient",
    expertId: input.expertId,
    expertName: input.expertName,
    topic: input.topic,
    status: "CONFIRMED",
    createdAt: new Date().toISOString(),
  };
  mockAppointments = [...mockAppointments, appointment];
  return appointment;
}

export function cancelMockAppointment(appointmentId: string): void {
  mockAppointments = mockAppointments.map((appointment) =>
    appointment.appointmentId === appointmentId
      ? { ...appointment, status: "CANCELED" }
      : appointment
  );
}
