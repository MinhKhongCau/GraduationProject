"use client";

import { useApiMutation } from "./useApiMutation";
import { profileApi } from "@/api";
import { AVATAR_ACCEPTED_TYPES, AVATAR_MAX_BYTES } from "@/api/profile";
import { useAuthContext } from "@/context/AuthContext";
import { useErrorContext } from "@/context/ErrorContext";

/**
 * Uploads the logged-in user's avatar to S3 (POST /profiles/:id/avatar) and
 * hands the returned S3 object key to `onUploaded`, so the form can save it as avatarUrl.
 */
export function useAvatarUpload(onUploaded: (objectKey: string) => void) {
  const { user } = useAuthContext();
  const { showError } = useErrorContext();

  const mutation = useApiMutation({
    mutationFn: (file: File) => {
      if (!user) throw new Error("You need to be logged in to upload an avatar.");
      return profileApi.uploadAvatar(user.id, file);
    },
    onSuccess: (result) => onUploaded(result.objectKey),
  });

  function upload(file: File) {
    if (!AVATAR_ACCEPTED_TYPES.includes(file.type)) {
      showError("Please choose a JPG, PNG, GIF or WebP image.");
      return;
    }
    if (file.size > AVATAR_MAX_BYTES) {
      showError("Image is too large (max 5MB).");
      return;
    }
    mutation.mutate(file);
  }

  return { upload, isUploading: mutation.isPending };
}
