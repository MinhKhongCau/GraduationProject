"use client";

import type { ReactNode } from "react";
import { GoogleOAuthProvider } from "@react-oauth/google";
import { QueryProvider } from "./QueryProvider";
import { LocaleProvider, type Locale } from "./LocaleContext";
import { AuthProvider } from "./AuthContext";
import { ChatProvider } from "./ChatContext";
import { ErrorProvider } from "./ErrorContext";
import { ToastContainer } from "@/components/ui/Toast";

/**
 * Google OAuth and the Capacitor Remote Server Shell:
 *
 * The native WebView loads the deployed HTTPS site, so Google sees the site's
 * real origin (for example `https://myapp.example.com`), not
 * `capacitor://localhost`. Add that exact HTTPS URL in Google Cloud Console →
 * Credentials → OAuth 2.0 Client ID → Authorized JavaScript origins. The
 * current GoogleLogin popup/callback flow has no redirect URI; if the backend
 * later uses an OAuth code redirect flow, add its exact callback URL under
 * Authorized redirect URIs as well.
 *
 * Do not add `capacitor://localhost` for this remote-shell model. Google policy
 * does not permit OAuth requests in an embedded user-agent, so the native app
 * must use native Google Sign-In rather than this web SDK when Google login is
 * enabled. `@capawesome/capacitor-google-sign-in` currently supports
 * Capacitor 8; keep the returned ID token exchange compatible with the future
 * POST /auth/google backend endpoint.
 */
const GOOGLE_CLIENT_ID = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;
const GOOGLE_AUTH_ENABLED = process.env.NEXT_PUBLIC_ENABLE_GOOGLE_AUTH === "true" && !!GOOGLE_CLIENT_ID;

function WithGoogleOAuth({ children }: { children: ReactNode }) {
  if (!GOOGLE_AUTH_ENABLED) return <>{children}</>;
  return <GoogleOAuthProvider clientId={GOOGLE_CLIENT_ID!}>{children}</GoogleOAuthProvider>;
}

export interface AppProvidersProps {
  children: ReactNode;
  initialLocale: Locale;
}

export function AppProviders({ children, initialLocale }: AppProvidersProps) {
  return (
    <WithGoogleOAuth>
      <QueryProvider>
        <LocaleProvider initialLocale={initialLocale}>
          <AuthProvider>
            <ChatProvider>
              <ErrorProvider>
                {children}
                <ToastContainer />
              </ErrorProvider>
            </ChatProvider>
          </AuthProvider>
        </LocaleProvider>
      </QueryProvider>
    </WithGoogleOAuth>
  );
}
