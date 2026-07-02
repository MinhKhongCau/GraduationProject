"use client";

import { useRouter } from "next/navigation";
import { useAuthContext } from "@/context/AuthContext";
import { useApiMutation } from "./useApiMutation";
import { authApi } from "@/api";
import { getRefreshToken } from "@/api/http/session";
import { ROUTES } from "@/constants/route";
import type { LoginRequest, LoginResponse, RegisterRequest } from "@/types";

export function useAuth() {
  const auth = useAuthContext();
  const router = useRouter();

  const loginMutation = useApiMutation({
    mutationFn: (payload: LoginRequest) => authApi.login(payload),
  });

  const registerMutation = useApiMutation({
    mutationFn: (payload: RegisterRequest) => authApi.register(payload),
  });

  function login(payload: LoginRequest, onSuccess?: (response: LoginResponse) => void) {
    loginMutation.mutate(payload, {
      onSuccess: (response) => {
        auth.applySession(response, payload.email);
        onSuccess?.(response);
      },
    });
  }

  function logout() {
    const refreshToken = getRefreshToken();
    auth.clearAuth();
    router.push(ROUTES.AUTH.LOGIN);
    if (refreshToken) {
      authApi.logout(refreshToken).catch(() => undefined);
    }
  }

  return { ...auth, login, loginMutation, registerMutation, logout };
}
