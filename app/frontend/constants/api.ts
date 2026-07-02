export const AUTH_ENDPOINTS = {
  REGISTER: "/auth/register",
  LOGIN: "/auth/login",
  REFRESH: "/auth/refresh",
  LOGOUT: "/auth/logout",
  ME: "/auth/me",
  PROFILE: "/auth/profile",
  CHANGE_PASSWORD: "/auth/change-password",
  /** Documented only — auth-service has no handler yet. */
  FORGOT_PASSWORD: "/auth/forgot-password",
  /** Documented only — auth-service has no handler yet. */
  RESET_PASSWORD: "/auth/reset-password",
  /** Documented only — auth-service has no handler yet. */
  GOOGLE: "/auth/google",
};

export const PROFILE_ENDPOINTS = {
  PATIENT: (accountId: string) => `/profiles/patients/${accountId}`,
  PATIENT_MEDICAL_HISTORIES: (accountId: string) =>
    `/profiles/patients/${accountId}/medical-histories`,
  SPECIALIZATIONS: "/profiles/specializations/",
  EXPERTS: "/profiles/experts/",
  EXPERT: (accountId: string) => `/profiles/experts/${accountId}`,
};

export const BOOKING_ENDPOINTS = {
  GENERATE_SLOTS: "/slots/generate",
  AVAILABLE_DATES: "/slots/available-dates",
  AVAILABLE_TIMES: "/slots/available-times",
  /** Documented only — booking-service has no handler yet, uses /data fallback. */
  LOCK: "/bookings/lock",
  /** Documented only — booking-service has no handler yet, uses /data fallback. */
  HISTORY: "/bookings/history",
  /** Documented only — booking-service has no handler yet, uses /data fallback. */
  WEEKLY_SCHEDULE: "/experts/me/schedule",
  /** Documented only — booking-service has no handler yet, uses /data fallback. */
  LEAVE_REQUESTS: "/experts/leave-requests",
};

export const PAYMENT_ENDPOINTS = {
  INIT_WALLET: "/payments/wallets/init",
  WALLET: (ownerId: string) => `/payments/wallets/${ownerId}`,
  TOP_UP: (ownerId: string) => `/payments/wallets/${ownerId}/top-up`,
  PAY: "/payments/wallets/pay",
  WITHDRAW: (ownerId: string) => `/payments/wallets/${ownerId}/withdraw`,
  PROCESS_WITHDRAWAL: (requestId: string) =>
    `/payments/wallets/withdrawals/${requestId}/process`,
};

export const ASSESSMENT_ENDPOINTS = {
  TEMPLATES: "/assessments/templates",
  QUESTIONS: (templateId: string) => `/assessments/templates/${templateId}/questions`,
  SUBMIT: "/assessments/submit",
};

/** Documented only — no service implements clinical records yet, uses /data fallback. */
export const CLINICAL_RECORD_ENDPOINTS = {
  CREATE: "/clinical-records",
  MY_HISTORY: "/clinical-records/my-history",
  CLIENT_HISTORY: (clientId: string) => `/clinical-records/client/${clientId}`,
};

export const STORAGE_KEYS = {
  ACCESS_TOKEN: "mc_access_token",
  REFRESH_TOKEN: "mc_refresh_token",
  ROLE: "mc_role",
  ACCOUNT_ID: "mc_account_id",
  FULL_NAME: "mc_full_name",
  EMAIL: "mc_email",
  LOCALE: "mc_locale",
} as const;

/** Cookies mirrored client-side purely for proxy.ts's coarse redirect — see DESIGN.md. */
export const SESSION_COOKIE = "mc_session";
export const ROLE_COOKIE = "mc_role";
export const LOCALE_COOKIE = "mc_locale";

export const QUERY_KEYS = {
  me: () => ["auth", "me"] as const,
  patientProfile: (accountId: string) => ["patient", "profile", accountId] as const,
  medicalHistories: (accountId: string) => ["patient", "medical-histories", accountId] as const,
  specializations: () => ["specializations"] as const,
  experts: (filters?: Record<string, unknown>) => ["experts", filters ?? {}] as const,
  expertProfile: (accountId: string) => ["expert", "profile", accountId] as const,
  availableDates: (expertId: string, month: number, year: number) =>
    ["booking", "available-dates", expertId, month, year] as const,
  availableTimes: (expertId: string, date: string) =>
    ["booking", "available-times", expertId, date] as const,
  bookingHistory: () => ["booking", "history"] as const,
  wallet: (ownerId: string) => ["payment", "wallet", ownerId] as const,
  assessmentTemplates: () => ["assessment", "templates"] as const,
  assessmentQuestions: (templateId: string) => ["assessment", "questions", templateId] as const,
  myMedicalRecordHistory: () => ["clinical-records", "my-history"] as const,
};
