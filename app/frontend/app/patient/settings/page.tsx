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
    queryKey: QUERY_KEYS.myPatientProfile(),
    queryFn: () => patientApi.getMyProfile(),
    enabled: !!user,
    retry: 0,
  });

  const updateProfileMutation = useApiMutation({
    mutationFn: async (values: ProfileFormValues) => {
      await authApi.updateProfile({
        fullName: values.fullName,
        dateOfBirth: values.dateOfBirth || undefined,
      });
      return patientApi.updateMyProfile({
        fullName: values.fullName,
        dateOfBirth: values.dateOfBirth || undefined,
        avatarUrl: values.avatarUrl,
      });
    },
    onSuccess: () => {
      showSuccess("Profile updated.");
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.myPatientProfile() });
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
          avatarUrl: profile?.avatarUrl ?? "",
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
