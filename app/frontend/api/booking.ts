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
  TimeTemplate,
  Availability,
  CreateAvailabilityRequest,
  UpdateAvailabilityRequest,
  GenerateSlotsResponse,
  GetExpertSlotsParams,
  ExpertSlot,
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

/** [PUBLIC] Admin-managed shift templates an expert can attach to a weekday. */
export async function getShiftTemplates(): Promise<TimeTemplate[]> {
  const response = await bookingClient.get<ServiceEnvelope<any>>(
    BOOKING_ENDPOINTS.SHIFT_TEMPLATES
  );
  const data = response.data.data;
  if (Array.isArray(data)) return data;
  return data?.templates ?? data?.items ?? [];
}

/** [EXPERT] The logged-in expert's own weekly template rows (one per enabled weekday). */
export async function getMyAvailabilities(): Promise<Availability[]> {
  const response = await bookingClient.get<ServiceEnvelope<any>>(
    BOOKING_ENDPOINTS.AVAILABILITIES
  );
  const data = response.data.data;
  if (Array.isArray(data)) return data;
  return data?.availabilities ?? data?.items ?? [];
}

export async function createAvailability(payload: CreateAvailabilityRequest): Promise<Availability> {
  const response = await bookingClient.post<ServiceEnvelope<Availability>>(
    BOOKING_ENDPOINTS.AVAILABILITIES,
    payload
  );
  return response.data.data;
}

export async function updateAvailability(
  availabilityId: string,
  payload: UpdateAvailabilityRequest
): Promise<void> {
  await bookingClient.patch(BOOKING_ENDPOINTS.AVAILABILITY(availabilityId), payload);
}

/** [EXPERT] Materializes concrete bookable slots for the next N days from the expert's weekly template. */
export async function generateSlots(daysToGenerate: number): Promise<GenerateSlotsResponse> {
  const response = await bookingClient.post<ServiceEnvelope<GenerateSlotsResponse>>(
    BOOKING_ENDPOINTS.GENERATE_SLOTS,
    { daysToGenerate }
  );
  return response.data.data;
}

/** [EXPERT] Own generated slots (AVAILABLE/LOCKED/OCCUPIED) for the calendar preview. */
export async function getExpertSlots(params: GetExpertSlotsParams = {}): Promise<ExpertSlot[]> {
  const response = await bookingClient.get<ServiceEnvelope<{ slots: ExpertSlot[]; total: number }>>(
    BOOKING_ENDPOINTS.EXPERT_SLOTS,
    { params }
  );
  return response.data.data.slots ?? [];
}

/** [EXPERT] Save or update medical record notes for an appointment. */
export async function saveMedicalRecord(
  appointmentId: string,
  payload: import("@/types").SaveMedicalRecordRequest
): Promise<import("@/types").MedicalRecord> {
  const response = await bookingClient.post<ServiceEnvelope<import("@/types").MedicalRecord>>(
    BOOKING_ENDPOINTS.APPOINTMENT_MEDICAL_RECORD(appointmentId),
    payload
  );
  return response.data.data;
}

/** [PATIENT/EXPERT] Get medical record for a specific appointment. */
export async function getAppointmentMedicalRecord(
  appointmentId: string
): Promise<import("@/types").MedicalRecord | null> {
  try {
    const response = await bookingClient.get<ServiceEnvelope<import("@/types").MedicalRecord>>(
      BOOKING_ENDPOINTS.APPOINTMENT_MEDICAL_RECORD(appointmentId)
    );
    return response.data.data ?? null;
  } catch {
    return null;
  }
}

/** [PATIENT/EXPERT] List medical records of current user. */
export async function getMedicalRecords(
  params: { page?: number; size?: number } = {}
): Promise<{ items: import("@/types").MedicalRecord[]; total: number }> {
  const response = await bookingClient.get<
    ServiceEnvelope<{ items?: import("@/types").MedicalRecord[]; total?: number; total_items?: number }>
  >(BOOKING_ENDPOINTS.MEDICAL_RECORDS, { params });
  const data = response.data.data;
  return {
    items: data.items ?? [],
    total: data.total ?? data.total_items ?? 0,
  };
}

/** [PATIENT/EXPERT] Get medical record detail by recordId. */
export async function getMedicalRecordDetail(
  recordId: string
): Promise<import("@/types").MedicalRecord> {
  const response = await bookingClient.get<ServiceEnvelope<import("@/types").MedicalRecord>>(
    BOOKING_ENDPOINTS.MEDICAL_RECORD(recordId)
  );
  return response.data.data;
}

