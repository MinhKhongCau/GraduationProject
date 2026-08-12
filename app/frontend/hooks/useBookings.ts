"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useApiQuery } from "./useApiQuery";
import { useApiMutation } from "./useApiMutation";
import { bookingApi, expertApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { Appointment, GetExpertAppointmentsParams } from "@/types";

export interface AppointmentWithExpert extends Appointment {
  expertName?: string;
}

/**
 * Appointment records carry only expertId — booking-service has no join to
 * profile-service — so every listing screen resolves display names itself.
 */
async function withExpertNames(appointments: Appointment[]): Promise<AppointmentWithExpert[]> {
  const expertIds = Array.from(new Set(appointments.map((appointment) => appointment.expertId)));
  const profiles = await Promise.all(
    expertIds.map((expertId) => expertApi.getExpertProfile(expertId).catch(() => null))
  );
  const nameByExpertId = new Map(expertIds.map((id, index) => [id, profiles[index]?.fullName]));

  return appointments.map((appointment) => ({
    ...appointment,
    expertName: nameByExpertId.get(appointment.expertId),
  }));
}

export interface AppointmentWithPatient extends Appointment {
  patientName?: string;
  patientEmail?: string;
}

async function withPatientNames(appointments: Appointment[]): Promise<AppointmentWithPatient[]> {
  const patientIds = Array.from(new Set(appointments.map((a) => a.patientId).filter(Boolean)));
  const profiles = await Promise.all(
    patientIds.map((patientId) => import("@/api").then((api) => api.patientApi.getPatientProfile(patientId)).catch(() => null))
  );
  const map = new Map(patientIds.map((id, index) => [id, profiles[index]]));

  return appointments.map((appointment) => {
    const profile = map.get(appointment.patientId);
    return {
      ...appointment,
      patientName: profile?.fullName,
      patientEmail: profile?.email,
    };
  });
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
