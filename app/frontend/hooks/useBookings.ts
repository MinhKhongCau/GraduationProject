"use client";

import { keepPreviousData, useQueries, useQueryClient } from "@tanstack/react-query";
import { useApiQuery } from "./useApiQuery";
import { useApiMutation } from "./useApiMutation";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type {
  Appointment,
  AppointmentListParams,
  AppointmentPage,
  AppointmentStatus,
  GetExpertAppointmentsParams,
} from "@/types";

/** Shares the prefix useCancelBooking invalidates, so cancelling refreshes the paged list and counts. */
const EXPERT_APPOINTMENTS_PREFIX = ["booking", "expert-appointments"] as const;

export interface AppointmentWithExpert extends Appointment {
  expertName?: string;
}

export interface AppointmentWithPatient extends Appointment {
  patientName?: string;
  patientEmail?: string;
}

/** booking-service attaches profiles (one batched gRPC call), so no per-row profile requests. */
function withExpertNames(appointments: Appointment[]): AppointmentWithExpert[] {
  return appointments.map((appointment) => ({ ...appointment, expertName: appointment.expert?.fullName }));
}

/** Prefer the examined person (record snapshot), falling back to the booking account. */
function withPatientNames(appointments: Appointment[]): AppointmentWithPatient[] {
  return appointments.map((appointment) => ({
    ...appointment,
    patientName: appointment.patient?.fullName || appointment.patientAccount?.fullName,
    patientEmail: appointment.patient?.email || appointment.patientAccount?.email,
  }));
}

/** [PATIENT] My Bookings — GET /booking/appointments. */
export function useMyBookings() {
  return useApiQuery({
    queryKey: QUERY_KEYS.myBookings(),
    queryFn: async () => withExpertNames(await bookingApi.getMyBookings()),
  });
}

/** [PATIENT] One server page of My Bookings; keeps the previous page on screen while the next loads. */
export function useMyBookingsPage(params: AppointmentListParams) {
  return useApiQuery({
    queryKey: [...QUERY_KEYS.myBookings(), "page", params],
    queryFn: async () => {
      const page = await bookingApi.listMyBookings(params);
      return { ...page, items: withExpertNames(page.items) };
    },
    placeholderData: keepPreviousData,
  });
}

/**
 * Totals per status (plus "ALL") within the list's date range: size=1 pages where only
 * `totalItems` is read. A status is missing while loading or if its request failed.
 */
function useAppointmentCounts(
  keyPrefix: readonly unknown[],
  fetchPage: (params: AppointmentListParams) => Promise<AppointmentPage>,
  statuses: readonly AppointmentStatus[],
  range: Pick<AppointmentListParams, "from" | "to">
) {
  const results = useQueries({
    queries: [undefined, ...statuses].map((status) => ({
      queryKey: [...keyPrefix, "count", status ?? "ALL", range],
      queryFn: () => fetchPage({ ...range, status, size: 1 }),
    })),
  });
  const counts: Partial<Record<AppointmentStatus | "ALL", number>> = {};
  results.forEach((result, index) => {
    if (result.data) counts[index === 0 ? "ALL" : statuses[index - 1]] = result.data.totalItems;
  });
  return counts;
}

/** [PATIENT] Booking totals per status for the My Bookings filter links. */
export function useMyBookingCounts(
  statuses: readonly AppointmentStatus[],
  range: Pick<AppointmentListParams, "from" | "to"> = {}
) {
  return useAppointmentCounts(QUERY_KEYS.myBookings(), bookingApi.listMyBookings, statuses, range);
}

/** [EXPERT] One server page of the expert's appointments; keeps the previous page on screen while the next loads. */
export function useExpertAppointmentsPage(params: AppointmentListParams) {
  return useApiQuery({
    queryKey: [...EXPERT_APPOINTMENTS_PREFIX, "page", params],
    queryFn: async () => {
      const page = await bookingApi.listExpertAppointments(params);
      return { ...page, items: withPatientNames(page.items) };
    },
    placeholderData: keepPreviousData,
  });
}

/** [EXPERT] Appointment totals per status for the Appointments filter links. */
export function useExpertAppointmentCounts(
  statuses: readonly AppointmentStatus[],
  range: Pick<AppointmentListParams, "from" | "to"> = {}
) {
  return useAppointmentCounts(EXPERT_APPOINTMENTS_PREFIX, bookingApi.listExpertAppointments, statuses, range);
}

/** [EXPERT] GET /booking/appointments/expert. */
export function useExpertAppointments(params: GetExpertAppointmentsParams = {}) {
  return useApiQuery({
    queryKey: QUERY_KEYS.expertAppointments(params as Record<string, unknown>),
    queryFn: async () => withPatientNames(await bookingApi.getExpertAppointments(params)),
  });
}

export function useCancelBooking() {
  const queryClient = useQueryClient();

  return useApiMutation({
    mutationFn: ({ appointmentId, reason }: { appointmentId: string; reason: string }) =>
      bookingApi.cancelAppointment(appointmentId, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.myBookings() });
      queryClient.invalidateQueries({ queryKey: ["booking", "expert-appointments"] });
    },
  });
}
