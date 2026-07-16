"use client";

import { useApiQuery } from "./useApiQuery";
import { bookingApi, expertApi } from "@/api";
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
 * [EXPERT] Chat contacts = patients from booking history. GAP: unlike
 * experts, profile-service has no public/expert-permitted endpoint to
 * resolve a patient's display name (GET /profiles/patients/:id is
 * ADMIN-only — see profile-service's routes/routes.go) so this falls back
 * to a placeholder label. DM addressing still works correctly since it's
 * accountId-keyed, not name-keyed. Follow-up: add a patient-lookup
 * endpoint scoped to a patient's assigned expert, or denormalize
 * patientName onto Appointment in booking-service.
 */
export function useExpertChatContacts() {
  return useApiQuery({
    queryKey: ["chat", "contacts", "expert"],
    queryFn: async (): Promise<ChatContact[]> => {
      const appointments = await bookingApi.getExpertAppointments();
      const patientIds = Array.from(new Set(appointments.map((a) => a.patientId)));
      return patientIds.map((id) => ({
        id,
        role: "PATIENT" as const,
        fullName: `Patient #${id.slice(0, 8)}`,
      }));
    },
  });
}
