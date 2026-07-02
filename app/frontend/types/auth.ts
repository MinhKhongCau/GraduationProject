export type UserRole = "PATIENT" | "EXPERT" | "ADMIN";

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  accountId: string;
  fullName: string;
  role: UserRole;
}

export interface RegisterRequest {
  fullName: string;
  email: string;
  password: string;
  confirmPassword: string;
  dateOfBirth: string;
  role: Extract<UserRole, "PATIENT" | "EXPERT">;
}

export interface TokenRefreshRequest {
  refreshToken: string;
}

export interface LogoutRequest {
  refreshToken: string;
}

export interface ProfileUpdateRequest {
  fullName: string;
  dateOfBirth: string;
}

export interface ChangePasswordRequest {
  oldPassword: string;
  newPassword: string;
  confirmNewPassword: string;
}

/** Documented in API-document.md; not implemented by auth-service yet. */
export interface ForgotPasswordRequest {
  email: string;
}

/** Documented in API-document.md; not implemented by auth-service yet. */
export interface ResetPasswordRequest {
  token: string;
  newPassword: string;
}

/** Documented in API-document.md as POST /auth/google; not implemented yet. */
export interface GoogleLoginRequest {
  idToken: string;
}
