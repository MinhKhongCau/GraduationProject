import { bookingClient } from "./http/instances";
import { BOOKING_ENDPOINTS } from "@/constants/api";
import type {
  AvailableDatesResponse,
  AvailableTimesResponse,
  AvailableTimeSlot,
  LockSlotResponse,
  CreateAppointmentRequest,
  CreateAppointmentResponse,
  BookingListResponse,
  GetExpertAppointmentsParams,
  Appointment,
  ServiceEnvelope,
} from "@/types";

export async function getAvailableDates(expertId: string): Promise<string[]> {
  const response = await bookingClient.get<ServiceEnvelope<AvailableDatesResponse>>(
    BOOKING_ENDPOINTS.AVAILABLE_DATES,
    { params: { expertId } }
  );
  return response.data.data.availableDates ?? [];
}

export async function getAvailableTimes(expertId: string, date: string): Promise<AvailableTimeSlot[]> {
  const response = await bookingClient.get<ServiceEnvelope<AvailableTimesResponse>>(
    BOOKING_ENDPOINTS.AVAILABLE_TIMES,
    { params: { expertId, date } }
  );
  return response.data.data.availableTimes ?? [];
}

/** Reserves a slot for 15 minutes so the patient can complete createAppointment. */
export async function lockSlot(slotId: string): Promise<LockSlotResponse> {
  const response = await bookingClient.post<ServiceEnvelope<LockSlotResponse>>(
    BOOKING_ENDPOINTS.LOCK_SLOT(slotId)
  );
  return response.data.data;
}

/** Must be called after lockSlot succeeds; leaves the appointment PENDING_PAYMENT. */
export async function createAppointment(
  payload: CreateAppointmentRequest
): Promise<CreateAppointmentResponse> {
  const response = await bookingClient.post<ServiceEnvelope<CreateAppointmentResponse>>(
    BOOKING_ENDPOINTS.APPOINTMENTS,
    payload
  );
  return response.data.data;
}

/** [PATIENT] GET /booking/appointments — no pagination/filter/sort support server-side. */
export async function getMyBookings(): Promise<Appointment[]> {
  const response = await bookingClient.get<ServiceEnvelope<BookingListResponse>>(
    BOOKING_ENDPOINTS.APPOINTMENTS
  );
  return response.data.data.appointments ?? [];
}

/** [EXPERT] GET /booking/appointments/expert */
export async function getExpertAppointments(
  params: GetExpertAppointmentsParams = {}
): Promise<Appointment[]> {
  const response = await bookingClient.get<ServiceEnvelope<BookingListResponse>>(
    BOOKING_ENDPOINTS.EXPERT_APPOINTMENTS,
    { params }
  );
  return response.data.data.appointments ?? [];
}

/** [PATIENT/EXPERT] `reason` is required by the backend. */
export async function cancelAppointment(appointmentId: string, reason: string): Promise<void> {
  await bookingClient.patch(BOOKING_ENDPOINTS.CANCEL_APPOINTMENT(appointmentId), { reason });
}
