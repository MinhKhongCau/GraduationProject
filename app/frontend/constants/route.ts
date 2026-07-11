export const ROUTES = {
  HOME: "/",
  FORBIDDEN: "/forbidden",

  AUTH: {
    LOGIN: "/auth/login",
    REGISTER: "/auth/register",
    FORGOT_PASSWORD: "/auth/forgot-password",
    RESET_PASSWORD: "/auth/reset-password",
  },

  PATIENT: {
    DASHBOARD: "/patient/dashboard",
    FIND_EXPERTS: "/patient/find-experts",
    EXPERT_DETAIL: (expertId: string) => `/patient/experts/${expertId}`,
    BOOK_APPOINTMENT: "/patient/book-appointment",
    MY_BOOKINGS: "/patient/my-bookings",
    WALLET: "/patient/wallet",
    ASSESSMENT: "/patient/assessment",
    ASSESSMENT_DETAIL: (templateId: string) => `/patient/assessment/${templateId}`,
    MEDICAL_HISTORY: "/patient/medical-history",
    MESSAGES: "/patient/messages",
    SETTINGS: "/patient/settings",
  },

  EXPERT: {
    DASHBOARD: "/expert/dashboard",
    SCHEDULE: "/expert/schedule",
    APPOINTMENTS: "/expert/appointments",
    PATIENTS: "/expert/patients",
    CLINICAL_RECORDS: "/expert/clinical-records",
    WALLET: "/expert/wallet",
    SETTINGS: "/expert/settings",
  },

  ADMIN: {
    DASHBOARD: "/admin/dashboard",
    EXPERTS: "/admin/experts",
    SPECIALIZATIONS: "/admin/specializations",
    PATIENTS: "/admin/patients",
    APPOINTMENTS: "/admin/appointments",
    SCHEDULES: "/admin/schedules",
    WITHDRAWALS: "/admin/withdrawals",
    TEMPLATES: "/admin/templates",
    DIMENSIONS: "/admin/dimensions",
    OPTION_GROUPS: "/admin/option-groups",
    QUESTIONS: "/admin/questions",
  },
} as const;
