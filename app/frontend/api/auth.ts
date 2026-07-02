import { authClient } from "./http/instances";
import { AUTH_ENDPOINTS } from "@/constants/api";
import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  MessageResponse,
  ProfileUpdateRequest,
  ChangePasswordRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  GoogleLoginRequest,
} from "@/types";

export async function login(payload: LoginRequest): Promise<LoginResponse> {
  const response = await authClient.post<LoginResponse>(AUTH_ENDPOINTS.LOGIN, payload);
  return response.data;
}

export async function register(payload: RegisterRequest): Promise<MessageResponse> {
  const response = await authClient.post<MessageResponse>(AUTH_ENDPOINTS.REGISTER, payload);
  return response.data;
}

export async function refreshToken(refreshTokenValue: string): Promise<LoginResponse> {
  const response = await authClient.post<LoginResponse>(AUTH_ENDPOINTS.REFRESH, {
    refreshToken: refreshTokenValue,
  });
  return response.data;
}

export async function logout(refreshTokenValue: string): Promise<void> {
  await authClient.post(AUTH_ENDPOINTS.LOGOUT, { refreshToken: refreshTokenValue });
}

export async function getMe(): Promise<MessageResponse> {
  const response = await authClient.get<MessageResponse>(AUTH_ENDPOINTS.ME);
  return response.data;
}

export async function updateProfile(payload: ProfileUpdateRequest): Promise<MessageResponse> {
  const response = await authClient.put<MessageResponse>(AUTH_ENDPOINTS.PROFILE, payload);
  return response.data;
}

export async function changePassword(payload: ChangePasswordRequest): Promise<MessageResponse> {
  const response = await authClient.put<MessageResponse>(AUTH_ENDPOINTS.CHANGE_PASSWORD, payload);
  return response.data;
}

/** Documented only — auth-service has no handler yet, expect this to fail gracefully. */
export async function forgotPassword(payload: ForgotPasswordRequest): Promise<MessageResponse> {
  const response = await authClient.post<MessageResponse>(AUTH_ENDPOINTS.FORGOT_PASSWORD, payload);
  return response.data;
}

/** Documented only — auth-service has no handler yet, expect this to fail gracefully. */
export async function resetPassword(payload: ResetPasswordRequest): Promise<MessageResponse> {
  const response = await authClient.post<MessageResponse>(AUTH_ENDPOINTS.RESET_PASSWORD, payload);
  return response.data;
}

/**
 * Documented only — auth-service has no /auth/google handler yet. Only
 * called when NEXT_PUBLIC_ENABLE_GOOGLE_AUTH="true"; callers must handle
 * the resulting 404 gracefully until the backend implements it.
 */
export async function loginWithGoogle(payload: GoogleLoginRequest): Promise<LoginResponse> {
  const response = await authClient.post<LoginResponse>(AUTH_ENDPOINTS.GOOGLE, payload);
  return response.data;
}
