"use client";

import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import {
  getAccessToken,
  getStoredUser,
  setSession,
  clearSession,
} from "@/api/http/session";
import type { LoginResponse, User } from "@/types";

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  applySession: (response: LoginResponse, email: string) => void;
  clearAuth: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // One-time hydration from localStorage after mount — must run in an
    // effect (not a lazy useState initializer) so the client's first render
    // matches the server-rendered HTML and avoids a hydration mismatch.
    const token = getAccessToken();
    const stored = getStoredUser();
    if (token && stored) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setUser(stored);
    }
    setIsLoading(false);
  }, []);

  const applySession = (response: LoginResponse, email: string) => {
    setSession({
      accessToken: response.accessToken,
      refreshToken: response.refreshToken,
      accountId: response.accountId,
      fullName: response.fullName,
      email,
      role: response.role,
    });
    setUser({ id: response.accountId, fullName: response.fullName, email, role: response.role });
  };

  const clearAuth = () => {
    clearSession();
    setUser(null);
  };

  const value = useMemo(
    () => ({ user, isLoading, isAuthenticated: !!user, applySession, clearAuth }),
    [user, isLoading]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuthContext(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuthContext must be used within an AuthProvider");
  }
  return context;
}
