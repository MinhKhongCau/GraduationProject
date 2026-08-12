"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useApiQuery } from "./useApiQuery";
import { useApiMutation } from "./useApiMutation";
import { bookingApi, expertApi, patientApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { MedicalRecord, SaveMedicalRecordRequest } from "@/types";

/**
 * Enriches medical records with patient and expert names by resolving profile IDs.
 */
async function enrichMedicalRecords(records: MedicalRecord[]): Promise<MedicalRecord[]> {
  const patientIds = Array.from(new Set(records.map((r) => r.patient_id || r.patientId || "").filter(Boolean)));
  const expertIds = Array.from(new Set(records.map((r) => r.expert_id || r.expertId || "").filter(Boolean)));

  const [patientProfiles, expertProfiles] = await Promise.all([
    Promise.all(patientIds.map((id) => patientApi.getPatientProfile(id).catch(() => null))),
    Promise.all(expertIds.map((id) => expertApi.getExpertProfile(id).catch(() => null))),
  ]);

  const patientMap = new Map(patientIds.map((id, index) => [id, patientProfiles[index]]));
  const expertMap = new Map(expertIds.map((id, index) => [id, expertProfiles[index]]));

  return records.map((record) => {
    const patId = record.patient_id || record.patientId || "";
    const expId = record.expert_id || record.expertId || "";
    const pat = patientMap.get(patId);
    const exp = expertMap.get(expId);

    return {
      ...record,
      patientName: pat?.fullName,
      patientEmail: pat?.email,
      patientAvatar: pat?.avatarUrl,
      expertName: exp?.fullName,
      expertAvatar: exp?.avatarUrl,
    };
  });
}

/** [PATIENT/EXPERT] Get all medical records of the logged in user. */
export function useMedicalRecords(params: { page?: number; size?: number } = {}) {
  return useApiQuery({
    queryKey: QUERY_KEYS.medicalRecords(params as Record<string, unknown>),
    queryFn: async () => {
      const res = await bookingApi.getMedicalRecords(params);
      const enriched = await enrichMedicalRecords(res.items);
      return {
        items: enriched,
        total: res.total,
      };
    },
  });
}

/** [PATIENT/EXPERT] Get medical record for a specific appointment. */
export function useAppointmentMedicalRecord(appointmentId?: string) {
  return useApiQuery({
    queryKey: appointmentId ? QUERY_KEYS.appointmentMedicalRecord(appointmentId) : ["booking", "none"],
    queryFn: async () => {
      if (!appointmentId) return null;
      const record = await bookingApi.getAppointmentMedicalRecord(appointmentId);
      if (!record) return null;
      const [enriched] = await enrichMedicalRecords([record]);
      return enriched ?? null;
    },
    enabled: !!appointmentId,
  });
}

/** [EXPERT] Mutation to create or update medical record for an appointment. */
export function useSaveMedicalRecord() {
  const queryClient = useQueryClient();

  return useApiMutation({
    mutationFn: ({
      appointmentId,
      payload,
    }: {
      appointmentId: string;
      payload: SaveMedicalRecordRequest;
    }) => bookingApi.saveMedicalRecord(appointmentId, payload),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.medicalRecords() });
      queryClient.invalidateQueries({
        queryKey: QUERY_KEYS.appointmentMedicalRecord(variables.appointmentId),
      });
      queryClient.invalidateQueries({ queryKey: ["booking", "expert-appointments"] });
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.myBookings() });
    },
  });
}
