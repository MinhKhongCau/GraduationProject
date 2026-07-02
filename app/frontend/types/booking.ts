export type SlotStatus = "AVAILABLE" | "UNAVAILABLE";

export interface ExpertSlot {
  slotId: string;
  expertId: string;
  dateSlot: string;
  startTime: string;
  endTime: string;
  status: SlotStatus;
  price: number;
  isLocked: boolean;
  lockedExpiresAt?: string;
}

export type AppointmentStatus =
  | "PENDING"
  | "LOCKED"
  | "CONFIRMED"
  | "CANCELED"
  | "COMPLETED";

export interface Appointment {
  appointmentId: string;
  slotId: string;
  patientId: string;
  expertId: string;
  expertName?: string;
  topic?: string;
  cancellationReason?: string;
  status: AppointmentStatus;
  meetingLink?: string;
  createdAt: string;
}

export interface AvailableDatesResponse {
  availableDates: string[];
}

export interface AvailableTimesResponse {
  date: string;
  timeSlots: ExpertSlot[];
}

/** Documented in API-document.md; booking-service has no handler yet. */
export interface LockAppointmentRequest {
  slotId: string;
}

/** Documented in API-document.md; booking-service has no handler yet. */
export interface BookingHistoryResponse {
  appointments: Appointment[];
}

export interface WeeklySlotInput {
  dayOfWeek:
    | "MONDAY"
    | "TUESDAY"
    | "WEDNESDAY"
    | "THURSDAY"
    | "FRIDAY"
    | "SATURDAY"
    | "SUNDAY";
  startTime: string;
  endTime: string;
}

/** Documented in API-document.md; booking-service has no handler yet. */
export interface WeeklyScheduleRequest {
  weeklySlots: WeeklySlotInput[];
}

export type LeaveRequestStatus = "PENDING" | "APPROVED" | "REJECTED";

/** Documented in API-document.md; booking-service has no handler yet. */
export interface LeaveRequest {
  startDate: string;
  endDate: string;
  reason: string;
  cancelAffectedBookings: boolean;
}

export interface LeaveRequestResponse {
  requestId: string;
  status: LeaveRequestStatus;
  affectedAppointments: number;
}
