"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ExpertProfileEditForm, type ExpertProfileFormValues } from "./component/ExpertProfileEditForm";
import { MySpecializationsCard } from "./component/MySpecializationsCard";
import { ChangePasswordForm } from "@/app/patient/settings/component/ChangePasswordForm";
import { useApiMutation, useApiQuery } from "@/hooks";
import { authApi, expertApi, specializationApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";
import { useErrorContext } from "@/context/ErrorContext";
import type { ChangePasswordRequest, ExpertProfile } from "@/types";
import { PageHeader } from "@/components/ui";

export default function ExpertSettingsPage() {
  const { user } = useAuthContext();
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();

  const { data: profile } = useQuery({
    queryKey: QUERY_KEYS.myExpertProfile(),
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
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.myExpertProfile() });
    },
  });

  const { data: activeSpecializations = [] } = useApiQuery({
    queryKey: QUERY_KEYS.specializations(),
    queryFn: () => specializationApi.getAllSpecializations(),
  });

  function onSpecializationsChanged(updated: ExpertProfile) {
    queryClient.setQueryData(QUERY_KEYS.myExpertProfile(), updated);
    // Public expert listings (booking, find-experts) filter by specialization.
    queryClient.invalidateQueries({ queryKey: ["experts"] });
  }

  const addSpecializationMutation = useApiMutation({
    mutationFn: (specId: string) => expertApi.addMySpecialization(specId),
    onSuccess: (updated) => {
      showSuccess("Specialization added.");
      onSpecializationsChanged(updated);
    },
  });

  const removeSpecializationMutation = useApiMutation({
    mutationFn: (specId: string) => expertApi.removeMySpecialization(specId),
    onSuccess: (updated) => {
      showSuccess("Specialization removed.");
      onSpecializationsChanged(updated);
    },
  });

  const changePasswordMutation = useApiMutation({
    mutationFn: (payload: ChangePasswordRequest) => authApi.changePassword(payload),
    onSuccess: () => showSuccess("Password updated."),
  });

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <PageHeader title="Settings" className="mb-0" />

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

      <MySpecializationsCard
        current={profile?.specializations ?? []}
        available={activeSpecializations}
        onAdd={(specId) => addSpecializationMutation.mutate(specId)}
        onRemove={(specId) => removeSpecializationMutation.mutate(specId)}
        isPending={addSpecializationMutation.isPending || removeSpecializationMutation.isPending}
      />

      <ChangePasswordForm
        onSubmit={(values) => changePasswordMutation.mutate(values)}
        isSubmitting={changePasswordMutation.isPending}
      />
    </div>
  );
}
