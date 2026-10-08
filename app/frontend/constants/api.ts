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
  ME: "/profiles/me",
  ME_MEDICAL_HISTORIES: "/profiles/me/medical-histories",
  ME_PATIENT_RECORDS: "/profiles/me/patient-records",
  PATIENTS: "/profiles/patients",
  PATIENT: (accountId: string) => `/profiles/patients/${accountId}`,
  PATIENT_PUBLIC: (accountId: string) => `/profiles/patients/${accountId}/public`,
  PATIENT_MEDICAL_HISTORIES: (accountId: string) =>
    `/profiles/patients/${accountId}/medical-histories`,
  SPECIALIZATIONS: "/profiles/specializations",
  SPECIALIZATIONS_ALL: "/profiles/specializations/all",
  SPECIALIZATION: (specId: string) => `/profiles/specializations/${specId}`,
  SPECIALIZATION_STATUS: (specId: string) => `/profiles/specializations/${specId}/status`,
  ME_SPECIALIZATION: (specId: string) => `/profiles/me/specializations/${specId}`,
  EXPERTS: "/profiles/experts",
  /** [ADMIN] Experts the logged-in admin approved and therefore manages. */
  MANAGED_EXPERTS: "/profiles/experts/managed",
  EXPERT: (accountId: string) => `/profiles/experts/${accountId}`,
  PROFILE: (accountId: string) => `/profiles/public/${accountId}`,
};

export const BOOKING_ENDPOINTS = {
  AVAILABLE_DATES: "/public/booking/slots/available-dates",
  AVAILABLE_TIMES: "/public/booking/slots/available-times",
  LOCK_SLOT: (slotId: string) => `/booking/slots/${slotId}/lock`,
  GENERATE_SLOTS: "/booking/slots/generate",
  EXPERT_SLOTS: "/booking/slots/expert",
  APPOINTMENTS: "/booking/appointments",
  EXPERT_APPOINTMENTS: "/booking/appointments/expert",
  APPOINTMENT_CONFIRMATION: "/booking/appointments/confirmation",
  CANCEL_APPOINTMENT: (appointmentId: string) => `/booking/appointments/${appointmentId}/cancel`,
  APPOINTMENT_MEDICAL_RECORD: (appointmentId: string) => `/booking/appointments/${appointmentId}/medical-record`,
  MEDICAL_RECORDS: "/booking/medical-records",
  MEDICAL_RECORD: (recordId: string) => `/booking/medical-records/${recordId}`,
  /** [PUBLIC] Admin-managed shift templates (e.g. "Morning shift 08:00-12:00"), expert picks from these. */
  SHIFT_TEMPLATES: "/public/booking/templates",
  /** [EXPERT] Weekly template — one row per weekday mapping a shift template to that day. */
  AVAILABILITIES: "/booking/availabilities",
  AVAILABILITY: (availabilityId: string) => `/booking/availabilities/${availabilityId}`,
  /** Documented only — booking-service has no handler yet, uses /data fallback. */
  LEAVE_REQUESTS: "/experts/leave-requests",
};

export const PAYMENT_ENDPOINTS = {
  INIT_WALLET: "/payments/wallets/me",
  WALLET: "/payments/wallets/me",
  TOP_UP: "/payments/wallets/top-up",
  PAY: "/payments/wallets/pay",
  WITHDRAW: "/payments/withdrawals",
  PROCESS_WITHDRAWAL: (requestId: string) =>
    `/payments/withdrawals/${requestId}/approve`,
  /** Creates a VNPay/MoMo order; when appointmentId is set, server recomputes amount from the real slot price. */
  ORDERS: "/payments/orders",
  /** Public: verifies the VNPay return query (vnp_SecureHash), settles the order and returns payment + booking status. */
  VNPAY_RETURN: "/payments/vnpay-return",
  /** [PATIENT] Totals of own payments (total paid counts SUCCESS only). */
  ORDERS_SUMMARY: "/payments/orders/summary",
  ORDER: (orderId: string) => `/payments/orders/${orderId}`,
  /** [EXPERT] Orders paid to me with gross / commission / net. */
  EXPERT_ORDERS: "/payments/expert/orders",
  EXPERT_ORDERS_SUMMARY: "/payments/expert/orders/summary",
  EXPERT_ORDER: (orderId: string) => `/payments/expert/orders/${orderId}`,
  /** [ADMIN] Scoped to experts the admin approved (= manages); others return 403. */
  ADMIN_ORDERS: "/payments/admin/orders",
  ADMIN_ORDERS_SUMMARY: "/payments/admin/orders/summary",
  ADMIN_ORDER: (orderId: string) => `/payments/admin/orders/${orderId}`,
  ADMIN_ORDER_REVIEW: (orderId: string) => `/payments/admin/orders/${orderId}/review`,
  ADMIN_WALLET_TRANSACTIONS: "/payments/admin/wallet-transactions",
  COMPENSATION_CASES: "/payments/compensation-cases",
  COMPENSATION_CASE_RESOLVE: (caseId: string) => `/payments/compensation-cases/${caseId}/resolve`,
};

export const ASSESSMENT_ENDPOINTS = {
  TEMPLATES: "/assessments/templates",
  TEMPLATE: (slug: string) => `/assessments/templates/${slug}`,
  QUESTIONS: (templateId: string) => `/assessments/templates/${templateId}/questions`,
  SUBMIT: "/assessments/submit",
  SELF: "/assessments/self",
  DIMENSIONS: "/assessments/dimensions",
  DIMENSION: (slug: string) => `/assessments/dimensions/${slug}`,
  DIMENSION_QUESTIONS_BULK: (slug: string) => `/assessments/dimensions/${slug}/questions/bulk`,
  OPTION_GROUPS: "/assessments/option-groups",
  OPTION_GROUP: (slug: string) => `/assessments/option-groups/${slug}`,
  QUESTION: (slug: string) => `/assessments/questions/${slug}`,
  BULK_QUESTIONS: "/assessments/questions/bulk",
};

export const FORUM_ENDPOINTS = {
  CATEGORIES: "/forum/categories",
  CATEGORY: (slug: string) => `/forum/categories/${slug}`,
  POSTS: "/forum/posts",
  /** GET treats :id as the SLUG; every other verb (POST/PUT/DELETE/PATCH,
   * and the comment/like/bookmark sub-routes) treats it as the numeric
   * post ID — a Gin radix-tree quirk, see forum-service's router.go. */
  POST: (slugOrId: string | number) => `/forum/posts/${slugOrId}`,
  POST_COMMENTS: (postId: number) => `/forum/posts/${postId}/comments`,
  POST_LIKE: (postId: number) => `/forum/posts/${postId}/like`,
  POST_BOOKMARK: (postId: number) => `/forum/posts/${postId}/bookmark`,
  POST_STATUS: (postId: number) => `/forum/posts/${postId}/status`,
  COMMENT: (commentId: number) => `/forum/comments/${commentId}`,
  TAGS: "/forum/tags",
  TAG_POSTS: (slug: string) => `/forum/tags/${slug}/posts`,
  MY_BOOKMARKS: "/forum/users/me/bookmarks",
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
  myPatientProfile: () => ["patient", "profile", "me"] as const,
  myMedicalHistories: () => ["patient", "medical-histories", "me"] as const,
  myPatientRecords: () => ["patient", "patient-records", "me"] as const,
  patients: (params?: Record<string, unknown>) => ["patients", params ?? {}] as const,
  patientProfile: (accountId: string) => ["patient", "profile", accountId] as const,
  patientMedicalHistories: (accountId: string) => ["patient", "medical-histories", accountId] as const,
  specializations: () => ["specializations"] as const,
  adminSpecializations: (params?: Record<string, unknown>) => ["specializations", "admin", params ?? {}] as const,
  myExpertProfile: () => ["expert", "profile", "me"] as const,
  experts: (filters?: Record<string, unknown>) => ["experts", filters ?? {}] as const,
  expertProfile: (accountId: string) => ["expert", "profile", accountId] as const,
  availableDates: (expertId: string) => ["booking", "available-dates", expertId] as const,
  availableTimes: (expertId: string, date: string) =>
    ["booking", "available-times", expertId, date] as const,
  myBookings: () => ["booking", "my-bookings"] as const,
  bookingConfirmation: (params: Record<string, unknown>) => ["booking", "confirmation", params] as const,
  expertAppointments: (filters?: Record<string, unknown>) =>
    ["booking", "expert-appointments", filters ?? {}] as const,
  shiftTemplates: () => ["booking", "shift-templates"] as const,
  myAvailabilities: () => ["booking", "availabilities", "me"] as const,
  expertSlots: (params?: Record<string, unknown>) => ["booking", "expert-slots", params ?? {}] as const,
  wallet: (ownerId: string) => ["payment", "wallet", ownerId] as const,
  myPaymentOrders: (filters?: object) => ["payment", "orders", "me", filters ?? {}] as const,
  myPaymentSummary: (filters?: object) => ["payment", "orders", "me", "summary", filters ?? {}] as const,
  expertPaymentOrders: (filters?: object) => ["payment", "orders", "expert", filters ?? {}] as const,
  expertPaymentSummary: (filters?: object) => ["payment", "orders", "expert", "summary", filters ?? {}] as const,
  adminPaymentOrders: (filters?: object) => ["payment", "orders", "admin", filters ?? {}] as const,
  adminPaymentSummary: (filters?: object) => ["payment", "orders", "admin", "summary", filters ?? {}] as const,
  adminPaymentOrder: (orderId: string) => ["payment", "orders", "admin", "detail", orderId] as const,
  compensationCases: (filters?: object) => ["payment", "compensation-cases", filters ?? {}] as const,
  managedWalletTransactions: (filters?: object) => ["payment", "wallet-transactions", "managed", filters ?? {}] as const,
  managedExperts: (filters?: object) => ["experts", "managed", filters ?? {}] as const,
  assessmentTemplates: () => ["assessment", "templates"] as const,
  assessmentTemplate: (slug: string) => ["assessment", "template", slug] as const,
  assessmentQuestions: (templateId: string) => ["assessment", "questions", templateId] as const,
  assessmentDimensions: () => ["assessment", "dimensions"] as const,
  assessmentOptionGroups: () => ["assessment", "optionGroups"] as const,
  assessmentOptionGroup: (slug: string) => ["assessment", "optionGroup", slug] as const,
  assessmentQuestion: (slug: string) => ["assessment", "question", slug] as const,
  assessmentHistory: () => ["assessment", "history"] as const,
  myMedicalRecordHistory: () => ["clinical-records", "my-history"] as const,
  medicalRecords: (params?: Record<string, unknown>) => ["booking", "medical-records", params ?? {}] as const,
  appointmentMedicalRecord: (appointmentId: string) => ["booking", "appointment-medical-record", appointmentId] as const,
  forumPosts: (params?: Record<string, unknown>) => ["forum", "posts", params ?? {}] as const,
  forumPost: (slug: string) => ["forum", "post", slug] as const,
  forumComments: (postId: number) => ["forum", "comments", postId] as const,
  forumCategories: () => ["forum", "categories"] as const,
  forumTags: () => ["forum", "tags"] as const,
  myBookmarks: () => ["forum", "my-bookmarks"] as const,
  myPosts: (userId?: string) => ["forum", "my-posts", userId ?? ""] as const,
};
