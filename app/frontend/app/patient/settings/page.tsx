"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ProfileEditForm, type ProfileFormValues } from "./component/ProfileEditForm";
import { ChangePasswordForm } from "./component/ChangePasswordForm";
import { DeleteAccountSection } from "./component/DeleteAccountSection";
import { useApiMutation } from "@/hooks";
import { authApi, patientApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";
import { useErrorContext } from "@/context/ErrorContext";
import type { ChangePasswordRequest } from "@/types";

export default function PatientSettingsPage() {
  const { user } = useAuthContext();
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();

  const { data: profile } = useQuery({
    queryKey: user ? QUERY_KEYS.patientProfile(user.id) : ["patient", "profile", "anonymous"],
    queryFn: () => patientApi.getPatientProfile(user!.id),
    enabled: !!user,
    retry: 0,
  });

  const updateProfileMutation = useApiMutation({
    mutationFn: async (values: ProfileFormValues) => {
      await authApi.updateProfile({ fullName: values.fullName, dateOfBirth: values.dateOfBirth });
      return patientApi.updatePatientProfile(user!.id, {
        fullName: values.fullName,
        dateOfBirth: values.dateOfBirth,
        phoneNumber: values.phoneNumber || undefined,
        gender: values.gender || undefined,
        address: values.address || undefined,
      });
    },
    onSuccess: () => {
      showSuccess("Profile updated.");
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.patientProfile(user!.id) });
    },
  });

  const changePasswordMutation = useApiMutation({
    mutationFn: (payload: ChangePasswordRequest) => authApi.changePassword(payload),
    onSuccess: () => showSuccess("Password updated."),
  });

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Settings</h1>

      <ProfileEditForm
        defaultValues={{
          fullName: user?.fullName ?? "",
          dateOfBirth: profile?.dateOfBirth ?? "",
          phoneNumber: profile?.phoneNumber ?? "",
          gender: profile?.gender ?? "",
          address: profile?.address ?? "",
        }}
        onSubmit={(values) => updateProfileMutation.mutate(values)}
        isSubmitting={updateProfileMutation.isPending}
      />

      <ChangePasswordForm
        onSubmit={(values) => changePasswordMutation.mutate(values)}
        isSubmitting={changePasswordMutation.isPending}
      />

      <DeleteAccountSection />
    </div>
  );
}
