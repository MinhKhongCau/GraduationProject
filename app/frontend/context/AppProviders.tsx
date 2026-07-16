"use client";

import type { ReactNode } from "react";
import { GoogleOAuthProvider } from "@react-oauth/google";
import { QueryProvider } from "./QueryProvider";
import { LocaleProvider, type Locale } from "./LocaleContext";
import { AuthProvider } from "./AuthContext";
import { ChatProvider } from "./ChatContext";
import { ErrorProvider } from "./ErrorContext";
import { ToastContainer } from "@/components/ui/Toast";

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
