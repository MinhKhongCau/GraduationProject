"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { PersonalInfoSection } from "./component/PersonalInfoSection";
import { MedicalHistoryList } from "./component/MedicalHistoryList";
import { AddMedicalHistoryDialog } from "./component/AddMedicalHistoryDialog";
import { Button, Spinner } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { patientApi } from "@/api";
import { QUERY_KEYS, ROUTES } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";
import { useErrorContext } from "@/context/ErrorContext";
import type { CreateMedicalHistoryRequest } from "@/types";

export default function MedicalHistoryPage() {
  const { user } = useAuthContext();
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();
  const [addOpen, setAddOpen] = useState(false);

  const {
    data: profile,
    isLoading,
    isError,
  } = useQuery({
    queryKey: QUERY_KEYS.myPatientProfile(),
    queryFn: () => patientApi.getMyProfile(),
    enabled: !!user,
    retry: 0,
  });

  const { data: histories = [] } = useApiQuery({
    queryKey: QUERY_KEYS.myMedicalHistories(),
    queryFn: () => patientApi.getMedicalHistories(),
    enabled: !!profile,
  });

  const addHistoryMutation = useApiMutation({
    mutationFn: (payload: CreateMedicalHistoryRequest) => patientApi.addMedicalHistory(payload),
    onSuccess: () => {
      showSuccess("Medical history entry added.");
      setAddOpen(false);
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.myMedicalHistories() });
    },
  });

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner className="h-6 w-6" />
      </div>
    );
  }

  if (isError || !profile) {
    return (
      <div className="mx-auto max-w-2xl rounded-3xl bg-surface p-10 text-center">
        <p className="mb-4 text-sm text-muted-foreground">
          Complete your profile in Settings before adding medical history.
        </p>
        <Link href={ROUTES.PATIENT.SETTINGS}>
          <Button>Go to settings</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl space-y-8">
      <h1 className="text-2xl font-bold text-foreground">Medical History</h1>

      <div>
        <h3 className="mb-4 text-sm font-bold text-foreground">Personal Information</h3>
        <PersonalInfoSection profile={profile} />
      </div>

      <MedicalHistoryList histories={histories} onAdd={() => setAddOpen(true)} />

      <AddMedicalHistoryDialog
        open={addOpen}
        onOpenChange={setAddOpen}
        onSubmit={(payload) => addHistoryMutation.mutate(payload)}
        isSubmitting={addHistoryMutation.isPending}
      />
    </div>
  );
}
