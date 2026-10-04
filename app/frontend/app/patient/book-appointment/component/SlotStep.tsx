"use client";

import { Check, Clock } from "lucide-react";
import { WeekCalendar } from "./WeekCalendar";
import { Button, PageHeader, Spinner } from "@/components/ui";
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
      <PageHeader title="Pick a date &amp; time" />

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
                  aria-pressed={selectedSlot?.slotId === slot.slotId}
                  className={`flex h-10 items-center gap-1.5 rounded-lg border px-4 text-sm font-semibold transition-colors ${
                    selectedSlot?.slotId === slot.slotId
                      ? "border-primary bg-primary text-white"
                      : "border-border-strong bg-background text-foreground hover:border-primary/40 hover:bg-surface"
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
