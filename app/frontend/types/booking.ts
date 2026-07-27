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
  price?: number;
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
  price?: number;
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

/** Reusable shift definition (e.g. "Morning shift 08:00-12:00"). Admin-managed, expert-readable. */
export interface TimeTemplate {
  templateId: string;
  shiftName: string;
  startTime: string; // "HH:MM"
  endTime: string; // "HH:MM"
  slotDurationMinutes: number;
  isActive: boolean;
}

/** 1=Monday ... 7=Sunday. */
export type DayOfWeek = 1 | 2 | 3 | 4 | 5 | 6 | 7;

/** One weekday of an expert's recurring weekly template — a full week is up to 7 of these. */
export interface Availability {
  availabilityId: string;
  expertId: string;
  templateId: string;
  dayOfWeek: DayOfWeek;
  isEnabled: boolean;
  effectiveFrom: number; // Unix ms
  effectiveUntil: number | null; // Unix ms
}

export interface CreateAvailabilityRequest {
  templateId: string;
  dayOfWeek: DayOfWeek;
  effectiveFrom: number;
  effectiveUntil?: number | null;
}

export interface UpdateAvailabilityRequest {
  templateId?: string;
  dayOfWeek?: DayOfWeek;
  isEnabled?: boolean;
  effectiveFrom?: number;
  effectiveUntil?: number | null;
}

export interface GenerateSlotsResponse {
  expertId: string;
  slotsCreated: number;
}

export interface GetExpertSlotsParams {
  fromDate?: number;
  toDate?: number;
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
