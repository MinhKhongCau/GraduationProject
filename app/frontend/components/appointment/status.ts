import type { BadgeTone } from "@/components/ui";
import type { AppointmentStatus } from "@/types";

export const APPOINTMENT_STATUS_TONES: Record<AppointmentStatus, BadgeTone> = {
  PENDING_PAYMENT: "warning",
  CONFIRMED: "success",
  CANCELLED: "danger",
  COMPLETED: "primary",
};

export const APPOINTMENT_STATUS_TEXT: Record<AppointmentStatus, string> = {
  PENDING_PAYMENT: "Chờ thanh toán",
  CONFIRMED: "Đã xác nhận",
  CANCELLED: "Đã hủy",
  COMPLETED: "Đã hoàn thành",
};

/** "08:00 – 09:00, 12/10/2026" from slot start/end (Unix ms). */
export function formatAppointmentTime(startTime?: number, endTime?: number): string {
  if (!startTime) return "—";
  const start = new Date(startTime);
  const time = (date: Date) => date.toLocaleTimeString("vi-VN", { hour: "2-digit", minute: "2-digit" });
  const range = endTime ? `${time(start)} – ${time(new Date(endTime))}` : time(start);
  return `${range}, ${start.toLocaleDateString("vi-VN")}`;
}
