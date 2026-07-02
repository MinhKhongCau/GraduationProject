import type { WeeklySlotInput } from "@/types";

/**
 * booking-service has no POST /experts/me/schedule handler yet (see
 * DESIGN.md). In-memory store so the expert schedule page feels real
 * within a session.
 */
let mockWeeklySlots: WeeklySlotInput[] = [
  { dayOfWeek: "MONDAY", startTime: "08:00", endTime: "12:00" },
  { dayOfWeek: "WEDNESDAY", startTime: "14:00", endTime: "18:00" },
];

export function getMockWeeklySchedule(): WeeklySlotInput[] {
  return [...mockWeeklySlots];
}

export function setMockWeeklySchedule(slots: WeeklySlotInput[]): void {
  mockWeeklySlots = [...slots];
}
