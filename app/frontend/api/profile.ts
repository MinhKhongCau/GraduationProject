import { profileClient } from "./http/instances";
import { PROFILE_ENDPOINTS } from "@/constants/api";

export interface UploadAvatarResult {
  authId: string;
  /** S3 object key of the new avatar; show it with resolveImageUrl() from @/lib/image. */
  objectKey: string;
}

/** Same limits profile-service enforces (avatar_handler.go / s3/client.go). */
export const AVATAR_MAX_BYTES = 5 * 1024 * 1024;
export const AVATAR_ACCEPTED_TYPES = ["image/jpeg", "image/png", "image/gif", "image/webp"];

/**
 * Uploads an avatar image to S3 via profile-service. The service stores the
 * returned object key (not a full URL) in the patient/expert profile's avatar_url.
 */
export async function uploadAvatar(accountId: string, file: File): Promise<UploadAvatarResult> {
  const formData = new FormData();
  formData.append("file", file);
  const response = await profileClient.post<UploadAvatarResult>(
    PROFILE_ENDPOINTS.AVATAR(accountId),
    formData,
    { timeout: 60000 } // S3 upload of up to 5 MB can exceed the default 15s.
  );
  return response.data;
}
