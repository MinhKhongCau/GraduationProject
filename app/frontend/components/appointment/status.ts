import type { BadgeTone } from "@/components/ui";
import type { Appointment, AppointmentSlot, AppointmentStatus, MeetingStatus } from "@/types";

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

export const MEETING_STATUS_TEXT: Record<MeetingStatus, string> = {
  too_early: "Chưa tới giờ vào phòng",
  open: "Phòng tư vấn đang mở",
  too_late: "Phòng tư vấn đã đóng",
};

export const MEETING_STATUS_TONES: Record<MeetingStatus, BadgeTone> = {
  too_early: "neutral",
  open: "success",
  too_late: "warning",
};

/** The booked slot, from `slot` or the flat start/end/price fields older responses carry. */
export function appointmentSlot(appointment: Appointment): AppointmentSlot | null {
  if (appointment.slot?.startTime) return appointment.slot;
  if (!appointment.startTime) return null;
  const endTime = appointment.endTime ?? appointment.startTime;
  return {
    slotId: appointment.slotId,
    startTime: appointment.startTime,
    endTime,
    durationMinutes: Math.max(0, Math.round((endTime - appointment.startTime) / 60_000)),
    price: appointment.price ?? 0,
  };
}

/** "Thứ Ba, 13/10/2026" for the slot day. */
export function formatSlotDate(startTime: number): string {
  return new Date(startTime).toLocaleDateString("vi-VN", {
    weekday: "long",
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}

/** "08:00 – 09:00" for the slot hours. */
export function formatSlotHours(startTime: number, endTime?: number): string {
  const time = (ms: number) => new Date(ms).toLocaleTimeString("vi-VN", { hour: "2-digit", minute: "2-digit" });
  return endTime && endTime > startTime ? `${time(startTime)} – ${time(endTime)}` : time(startTime);
}
