import { bookingClient } from "./http/instances";
import { BOOKING_ENDPOINTS } from "@/constants/api";
import {
  listMockAppointments,
  addMockAppointment,
  cancelMockAppointment,
  getNextQueueNumber,
} from "@/data/booking-history";
import type { AvailableDatesResponse, AvailableTimesResponse, Appointment } from "@/types";

export async function getAvailableDates(
  expertId: string,
  month: number,
  year: number
): Promise<string[]> {
  const response = await bookingClient.get<AvailableDatesResponse>(
    BOOKING_ENDPOINTS.AVAILABLE_DATES,
    { params: { expertId, month, year } }
  );
  return response.data.availableDates ?? [];
}

export async function getAvailableTimes(
  expertId: string,
  date: string
): Promise<AvailableTimesResponse> {
  const response = await bookingClient.get<AvailableTimesResponse>(
    BOOKING_ENDPOINTS.AVAILABLE_TIMES,
    { params: { expertId, date } }
  );
  return response.data;
}

/**
 * Documented as POST /bookings/lock in API-document.md; booking-service has
 * no handler yet, so this confirms against the in-memory mock instead.
 */
export interface ConfirmBookingInput {
  slotId: string;
  expertId: string;
  expertName: string;
  topic: string;
}

export function confirmBookingMock(input: ConfirmBookingInput): Appointment {
  return addMockAppointment(input);
}

export function getBookingQueueNumber(slotId: string): number {
  return getNextQueueNumber(slotId);
}

/** Documented as GET /bookings/history; booking-service has no handler yet. */
export async function getBookingHistory(): Promise<Appointment[]> {
  return Promise.resolve(listMockAppointments());
}

/** No ownership-check bug carried over from the legacy PHP app — see DESIGN.md. */
export async function cancelAppointment(appointmentId: string): Promise<void> {
  cancelMockAppointment(appointmentId);
  return Promise.resolve();
}
