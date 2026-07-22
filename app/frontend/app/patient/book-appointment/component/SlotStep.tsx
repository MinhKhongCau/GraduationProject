"use client";

import { Check, Clock } from "lucide-react";
import { WeekCalendar } from "./WeekCalendar";
import { Button, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { AvailableTimeSlot } from "@/types";

export interface SlotStepProps {
  expertId: string;
  selectedDate: string | null;
  onSelectDate: (date: string) => void;
  selectedSlot: AvailableTimeSlot | null;
  onSelectSlot: (slot: AvailableTimeSlot) => void;
  onBack: () => void;
  onNext: () => void;
}

export function SlotStep({
  expertId,
  selectedDate,
  onSelectDate,
  selectedSlot,
  onSelectSlot,
  onBack,
  onNext,
}: SlotStepProps) {
  const { data: availableDates = [], isLoading: isLoadingDates } = useApiQuery({
    queryKey: QUERY_KEYS.availableDates(expertId),
    queryFn: () => bookingApi.getAvailableDates(expertId),
  });

  const { data: availableTimes = [], isLoading: isLoadingTimes } = useApiQuery({
    queryKey: QUERY_KEYS.availableTimes(expertId, selectedDate ?? ""),
    queryFn: () => bookingApi.getAvailableTimes(expertId, selectedDate!),
    enabled: !!selectedDate,
  });

  return (
    <div className="flex h-full flex-col">
      <h1 className="mb-6 text-2xl font-bold text-foreground">Pick a date &amp; time</h1>

      <div className="mb-6">
        <p className="mb-3 text-sm font-semibold text-foreground">Pick a day</p>
        <WeekCalendar
          availableDates={availableDates}
          selectedDate={selectedDate}
          onSelectDate={onSelectDate}
          isLoading={isLoadingDates}
        />
      </div>

      {selectedDate && (
        <div className="mb-8">
          <p className="mb-3 text-sm font-semibold text-foreground">Available times</p>
          {isLoadingTimes ? (
            <Spinner className="h-5 w-5" />
          ) : availableTimes.length === 0 ? (
            <p className="text-sm text-muted-foreground">No open slots on this date.</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {availableTimes.map((slot) => (
                <button
                  key={slot.slotId}
                  type="button"
                  onClick={() => onSelectSlot(slot)}
                  className={`flex items-center gap-1.5 rounded-xl border px-4 py-2 text-sm font-semibold transition-colors ${
                    selectedSlot?.slotId === slot.slotId
                      ? "border-primary bg-primary text-white"
                      : "border-border text-foreground hover:bg-surface"
                  }`}
                >
                  <Clock className="h-3.5 w-3.5" />
                  {new Date(slot.startTime).toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" })}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      <div className="mt-auto flex justify-between border-t border-border pt-6">
        <Button variant="outline" onClick={onBack}>
          Back
        </Button>
        <Button onClick={onNext} disabled={!selectedSlot}>
          Next
          <Check className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
