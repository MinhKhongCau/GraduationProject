"use client";

import { useState } from "react";
import { Button, Card } from "@/components/ui";
import type { WeeklySlotInput } from "@/types";

const DAYS: WeeklySlotInput["dayOfWeek"][] = [
  "MONDAY",
  "TUESDAY",
  "WEDNESDAY",
  "THURSDAY",
  "FRIDAY",
  "SATURDAY",
  "SUNDAY",
];

export interface WeeklyScheduleEditorProps {
  initialSlots: WeeklySlotInput[];
  onSave: (slots: WeeklySlotInput[]) => void;
  isSubmitting: boolean;
}

export function WeeklyScheduleEditor({ initialSlots, onSave, isSubmitting }: WeeklyScheduleEditorProps) {
  const [slots, setSlots] = useState<Record<string, { enabled: boolean; startTime: string; endTime: string }>>(() => {
    const map: Record<string, { enabled: boolean; startTime: string; endTime: string }> = {};
    for (const day of DAYS) {
      const existing = initialSlots.find((slot) => slot.dayOfWeek === day);
      map[day] = existing
        ? { enabled: true, startTime: existing.startTime, endTime: existing.endTime }
        : { enabled: false, startTime: "08:00", endTime: "12:00" };
    }
    return map;
  });

  function updateDay(day: string, changes: Partial<{ enabled: boolean; startTime: string; endTime: string }>) {
    setSlots((current) => ({ ...current, [day]: { ...current[day], ...changes } }));
  }

  function handleSave() {
    const weeklySlots: WeeklySlotInput[] = DAYS.filter((day) => slots[day].enabled).map((day) => ({
      dayOfWeek: day,
      startTime: slots[day].startTime,
      endTime: slots[day].endTime,
    }));
    onSave(weeklySlots);
  }

  return (
    <Card className="p-6">
      <div className="space-y-3">
        {DAYS.map((day) => (
          <div key={day} className="flex flex-col gap-2 border-b border-border pb-3 last:border-0 sm:flex-row sm:items-center">
            <label className="flex w-40 items-center gap-2 text-sm font-semibold text-foreground">
              <input
                type="checkbox"
                checked={slots[day].enabled}
                onChange={(event) => updateDay(day, { enabled: event.target.checked })}
              />
              {day.charAt(0) + day.slice(1).toLowerCase()}
            </label>
            {slots[day].enabled && (
              <div className="flex items-center gap-2">
                <input
                  type="time"
                  value={slots[day].startTime}
                  onChange={(event) => updateDay(day, { startTime: event.target.value })}
                  className="rounded-lg border border-border px-2 py-1.5 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
                />
                <span className="text-muted-foreground">to</span>
                <input
                  type="time"
                  value={slots[day].endTime}
                  onChange={(event) => updateDay(day, { endTime: event.target.value })}
                  className="rounded-lg border border-border px-2 py-1.5 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
                />
              </div>
            )}
          </div>
        ))}
      </div>
      <Button className="mt-6" onClick={handleSave} disabled={isSubmitting}>
        {isSubmitting ? "Saving..." : "Save schedule"}
      </Button>
    </Card>
  );
}
