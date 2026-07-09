export type SlotStatusCode = 0 | 1 | 2;
export type SlotStatus = "AVAILABLE" | "LOCKED" | "OCCUPIED";

/** Full slot record, as returned by GET /booking/slots/expert. */
export interface ExpertSlot {
  slotId: string;
  expertId: string;
  dateSlot: string;
  startTime: number;
  endTime: number;
  status: SlotStatusCode;
  statusLabel: SlotStatus;
  price: number;
  lockedExpiresAt: number | null;
  lockedBy: string | null;
  availabilityId: string | null;
  createdAt: number;
  updatedAt: number;
}

/** Lightweight slot shape returned by the public available-times endpoint. */
export interface AvailableTimeSlot {
  slotId: string;
  startTime: number;
  endTime: number;
}

export type AppointmentStatusCode = 0 | 1 | 2;
export type AppointmentStatus = "PENDING_PAYMENT" | "CONFIRMED" | "CANCELLED";

export interface Appointment {
  appointmentId: string;
  slotId: string;
  patientId: string;
  expertId: string;
  cancellationReason: string;
  cancelledBy: "SYSTEM" | "PATIENT" | "EXPERT" | null;
  status: AppointmentStatusCode;
  statusLabel: AppointmentStatus;
  meetingLink: string;
  createdAt: number;
  updatedAt: number;
  confirmedAt: number | null;
}

export interface AvailableDatesResponse {
  expertId: string;
  availableDates: string[];
}

export interface AvailableTimesResponse {
  expertId: string;
  date: string;
  availableTimes: AvailableTimeSlot[];
}

export interface LockSlotResponse {
  slotId: string;
  expires: number;
}

export interface CreateAppointmentRequest {
  slotId: string;
  expertId: string;
}

export interface CreateAppointmentResponse {
  appointmentId: string;
  status: AppointmentStatusCode;
  slotId: string;
}

export interface BookingListResponse {
  appointments: Appointment[];
  total: number;
}

export interface GetExpertAppointmentsParams {
  fromDate?: number;
  toDate?: number;
  status?: AppointmentStatusCode;
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

/** Documented in API-document.md; booking-service models this as templates + availabilities instead — no handler yet for this shape. */
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
