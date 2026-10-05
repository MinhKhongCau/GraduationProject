"use client";

import { useMemo, useState } from "react";
import { Calendar, type DateObject } from "react-multi-date-picker";
import { Clock } from "lucide-react";
import { Badge, Card, Spinner, type BadgeTone } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { ExpertSlot, SlotStatus } from "@/types";

const STATUS_TONES: Record<SlotStatus, BadgeTone> = {
  AVAILABLE: "success",
  LOCKED: "warning",
  OCCUPIED: "danger",
};

function toDateKey(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

/** Month calendar of the expert's own generated slots, backed by GET /booking/slots/expert. */
export function SlotCalendarPreview() {
  const [selectedDate, setSelectedDate] = useState<Date>(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return today;
  });

  const rangeParams = useMemo(() => {
    const from = new Date();
    from.setHours(0, 0, 0, 0);
    const to = new Date(from);
    to.setDate(to.getDate() + 60);
    return { fromDate: from.getTime(), toDate: to.getTime() };
  }, []);

  const { data: slots = [], isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.expertSlots(rangeParams),
    queryFn: () => bookingApi.getExpertSlots(rangeParams),
  });

  const slotsByDate = useMemo(() => {
    const map = new Map<string, ExpertSlot[]>();
    for (const slot of slots) {
      const key = toDateKey(new Date(slot.startTime));
      const list = map.get(key) ?? [];
      list.push(slot);
      map.set(key, list);
    }
    return map;
  }, [slots]);

  const daySlots = [...(slotsByDate.get(toDateKey(selectedDate)) ?? [])].sort(
    (a, b) => a.startTime - b.startTime
  );

  return (
    <Card className="p-5 sm:p-6">
      <h2 className="mb-1 text-base font-semibold text-foreground">Your slot calendar</h2>
      <p className="mb-4 text-sm text-muted-foreground">
        Bold days already have generated slots. Select a day to see its slots.
      </p>

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : (
        <div className="flex flex-col gap-6 sm:flex-row">
          <Calendar
            value={selectedDate}
            onChange={(date) => {
              const native = (date as DateObject | null)?.toDate();
              if (native) setSelectedDate(native);
            }}
            mapDays={({ date }) =>
              slotsByDate.has(toDateKey(date.toDate())) ? { style: { fontWeight: 800 } } : {}
            }
          />

          <div className="flex-1">
            <p className="mb-3 text-sm font-semibold text-foreground">
              {selectedDate.toLocaleDateString("en-GB", { weekday: "long", day: "2-digit", month: "short" })}
            </p>
            {daySlots.length === 0 ? (
              <p className="text-sm text-muted-foreground">No slots generated for this day.</p>
            ) : (
              <div className="space-y-2">
                {daySlots.map((slot) => (
                  <div
                    key={slot.slotId}
                    className="flex items-center justify-between rounded-lg border border-border bg-background px-3 py-2.5 text-sm"
                  >
                    <span className="flex items-center gap-1.5 text-foreground">
                      <Clock className="h-3.5 w-3.5" />
                      {new Date(slot.startTime).toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" })}
                    </span>
                    <Badge tone={STATUS_TONES[slot.statusLabel]} className="uppercase">
                      {slot.statusLabel}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </Card>
  );
}
