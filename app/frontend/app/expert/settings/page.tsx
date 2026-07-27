"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ExpertProfileEditForm, type ExpertProfileFormValues } from "./component/ExpertProfileEditForm";
import { ChangePasswordForm } from "@/app/patient/settings/component/ChangePasswordForm";
import { useApiMutation } from "@/hooks";
import { authApi, expertApi } from "@/api";
import { useAuthContext } from "@/context/AuthContext";
import { useErrorContext } from "@/context/ErrorContext";
import type { ChangePasswordRequest } from "@/types";

export default function ExpertSettingsPage() {
  const { user } = useAuthContext();
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();

  const { data: profile } = useQuery({
    queryKey: ["expert", "profile", "me"],
    queryFn: () => expertApi.getMyProfile(),
    enabled: !!user,
    retry: 0,
  });

  const updateProfileMutation = useApiMutation({
    mutationFn: async (values: ExpertProfileFormValues) => {
      await authApi.updateProfile({ fullName: values.fullName });
      return expertApi.updateMyProfile({
        fullName: values.fullName,
        avatarUrl: values.avatarUrl,
        introductionVideoUrl: values.introductionVideoUrl,
        bio: values.bio,
      });
    },
    onSuccess: () => {
      showSuccess("Profile updated.");
      queryClient.invalidateQueries({ queryKey: ["expert", "profile", "me"] });
    },
  });

  const changePasswordMutation = useApiMutation({
    mutationFn: (payload: ChangePasswordRequest) => authApi.changePassword(payload),
    onSuccess: () => showSuccess("Password updated."),
  });

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Settings</h1>

      <ExpertProfileEditForm
        defaultValues={{
          fullName: user?.fullName ?? "",
          avatarUrl: profile?.avatarUrl ?? "",
          introductionVideoUrl: profile?.introductionVideoUrl ?? "",
          bio: profile?.bio ?? "",
        }}
        onSubmit={(values) => updateProfileMutation.mutate(values)}
        isSubmitting={updateProfileMutation.isPending}
      />

      <ChangePasswordForm
        onSubmit={(values) => changePasswordMutation.mutate(values)}
        isSubmitting={changePasswordMutation.isPending}
      />
    </div>
  );
}
