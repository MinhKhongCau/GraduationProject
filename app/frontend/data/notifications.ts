/** Documented as GET /notifications in API-document.md; no service implements it yet. */
export interface MockNotification {
  id: string;
  title: string;
  body: string;
  read: boolean;
  createdAt: string;
}

export const NOTIFICATIONS_MOCK: MockNotification[] = [
  {
    id: "notif-1",
    title: "Appointment confirmed",
    body: "Your session with Dr. Tran Minh Tam is confirmed.",
    read: false,
    createdAt: new Date(2026, 5, 20, 9, 5).toISOString(),
  },
  {
    id: "notif-2",
    title: "Wallet top-up successful",
    body: "Your wallet has been topped up.",
    read: true,
    createdAt: new Date(2026, 5, 19, 14, 20).toISOString(),
  },
];
