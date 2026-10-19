/** Public base URL of the image bucket, e.g. https://pub-xxxx.r2.dev (no trailing slash). */
const S3_PUBLIC_URL = (process.env.NEXT_PUBLIC_S3_PUBLIC_URL ?? "").replace(/\/+$/, "");

/**
 * The backend stores S3 object keys (e.g. "profiles/<id>/avatar-<uuid>.png"), not full URLs.
 * This turns a key into a displayable URL by prefixing NEXT_PUBLIC_S3_PUBLIC_URL.
 * Values that are already URLs (old rows, mock data, local assets, blob/data previews) are returned as-is.
 */
export function resolveImageUrl(keyOrUrl?: string | null): string | undefined {
  if (!keyOrUrl) return undefined;
  if (/^(https?:|data:|blob:|\/)/i.test(keyOrUrl)) return keyOrUrl;
  return `${S3_PUBLIC_URL}/${keyOrUrl.replace(/^\/+/, "")}`;
}
