"use client";

import { useApiQuery } from "./useApiQuery";
import { bookingApi, expertApi, patientApi } from "@/api";
import type { ChatContact } from "@/types";

/**
 * [PATIENT] Chat contacts = experts from booking history — there's no
 * dedicated "my experts" endpoint, same derivation pattern as
 * useBookings.ts's withExpertNames(). Expert profiles are public so real
 * names/avatars resolve correctly.
 */
export function usePatientChatContacts() {
  return useApiQuery({
    queryKey: ["chat", "contacts", "patient"],
    queryFn: async (): Promise<ChatContact[]> => {
      const appointments = await bookingApi.getMyBookings();
      const expertIds = Array.from(new Set(appointments.map((a) => a.expertId)));
      const profiles = await Promise.all(
        expertIds.map((id) => expertApi.getExpertProfile(id).catch(() => null))
      );
      return expertIds.map((id, i) => ({
        id,
        role: "EXPERT" as const,
        fullName: profiles[i]?.fullName ?? "Expert",
        avatarUrl: profiles[i]?.avatarUrl,
      }));
    },
  });
}

/**
 * [EXPERT] Chat contacts = patients from booking history.
 * We resolve patient public details (name, avatar) using the public patient lookup.
 */
export function useExpertChatContacts() {
  return useApiQuery({
    queryKey: ["chat", "contacts", "expert"],
    queryFn: async (): Promise<ChatContact[]> => {
      const appointments = await bookingApi.getExpertAppointments();
      const patientIds = Array.from(new Set(appointments.map((a) => a.patientId)));
      const profiles = await Promise.all(
        patientIds.map((id) => patientApi.getPatientProfilePublic(id).catch(() => null))
      );
      return patientIds.map((id, i) => ({
        id,
        role: "PATIENT" as const,
        fullName: profiles[i]?.fullName ?? `Patient #${id.slice(0, 8)}`,
        avatarUrl: profiles[i]?.avatarUrl,
      }));
    },
  });
}
