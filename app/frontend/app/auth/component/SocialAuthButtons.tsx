"use client";

import { GoogleLogin } from "@react-oauth/google";
import { useErrorContext } from "@/context/ErrorContext";

const GOOGLE_AUTH_ENABLED = process.env.NEXT_PUBLIC_ENABLE_GOOGLE_AUTH === "true";

export interface SocialAuthButtonsProps {
  onGoogleCredential: (idToken: string) => void;
}

/**
 * Google login is feature-flagged: the backend has no POST /auth/google
 * handler yet (only documented in API-document.md — see DESIGN.md), so
 * this stays behind NEXT_PUBLIC_ENABLE_GOOGLE_AUTH until it's implemented.
 * Facebook stays UI-only per the original prototype — no backend contract
 * exists for it at all.
 */
export function SocialAuthButtons({ onGoogleCredential }: SocialAuthButtonsProps) {
  const { showError } = useErrorContext();

  return (
    <div className="space-y-3">
      <div className="relative flex items-center py-1">
        <div className="flex-grow border-t border-border" />
        <span className="mx-4 shrink-0 text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
          Or continue with
        </span>
        <div className="flex-grow border-t border-border" />
      </div>

      <div className="grid grid-cols-2 gap-3">
        {GOOGLE_AUTH_ENABLED ? (
          <div className="col-span-2 flex justify-center">
            <GoogleLogin
              onSuccess={(credential) => {
                if (credential.credential) onGoogleCredential(credential.credential);
              }}
              onError={() => showError({ message: "Google sign-in failed. Please try again." })}
              width="100%"
            />
          </div>
        ) : (
          <button
            type="button"
            disabled
            title="Google sign-in is not available yet"
            className="flex cursor-not-allowed items-center justify-center gap-2 rounded-xl border border-border py-2 text-sm font-semibold text-muted-foreground opacity-60"
          >
            Google
          </button>
        )}
        <button
          type="button"
          disabled
          title="Facebook sign-in is not available yet"
          className="flex cursor-not-allowed items-center justify-center gap-2 rounded-xl border border-border py-2 text-sm font-semibold text-muted-foreground opacity-60"
        >
          Facebook
        </button>
      </div>
    </div>
  );
}
