"use client";

import { keepPreviousData, useQueries, useQueryClient } from "@tanstack/react-query";
import { useApiQuery } from "./useApiQuery";
import { useApiMutation } from "./useApiMutation";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { Appointment, AppointmentListParams, AppointmentStatus, GetExpertAppointmentsParams } from "@/types";

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
 * [PATIENT] Booking totals per status within the same date range as the list (size=1 pages,
 * only `totalItems` is read). Missing while loading/failed.
 */
export function useMyBookingCounts(statuses: readonly AppointmentStatus[], range: Pick<AppointmentListParams, "from" | "to"> = {}) {
  const results = useQueries({
    queries: [undefined, ...statuses].map((status) => ({
      queryKey: [...QUERY_KEYS.myBookings(), "count", status ?? "ALL", range],
      queryFn: () => bookingApi.listMyBookings({ ...range, status, size: 1 }),
    })),
  });
  const counts: Partial<Record<AppointmentStatus | "ALL", number>> = {};
  results.forEach((result, index) => {
    if (result.data) counts[index === 0 ? "ALL" : statuses[index - 1]] = result.data.totalItems;
  });
  return counts;
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
