"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useApiQuery } from "./useApiQuery";
import { useApiMutation } from "./useApiMutation";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { Appointment, GetExpertAppointmentsParams } from "@/types";

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
