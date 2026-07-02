"use client";

import { useState } from "react";
import { Check, Clock } from "lucide-react";
import { Button, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { ExpertSlot } from "@/types";

export interface SlotStepProps {
  expertId: string;
  selectedSlot: ExpertSlot | null;
  onSelectSlot: (slot: ExpertSlot) => void;
  onBack: () => void;
  onNext: () => void;
}

export function SlotStep({ expertId, selectedSlot, onSelectSlot, onBack, onNext }: SlotStepProps) {
  const now = new Date();
  const [selectedDate, setSelectedDate] = useState<string | null>(null);

  const { data: availableDates = [], isLoading: isLoadingDates } = useApiQuery({
    queryKey: QUERY_KEYS.availableDates(expertId, now.getMonth() + 1, now.getFullYear()),
    queryFn: () => bookingApi.getAvailableDates(expertId, now.getMonth() + 1, now.getFullYear()),
  });

  const { data: timesResponse, isLoading: isLoadingTimes } = useApiQuery({
    queryKey: QUERY_KEYS.availableTimes(expertId, selectedDate ?? ""),
    queryFn: () => bookingApi.getAvailableTimes(expertId, selectedDate!),
    enabled: !!selectedDate,
  });

  return (
    <div className="flex h-full flex-col">
      <h1 className="mb-6 text-2xl font-bold text-foreground">Pick a date &amp; time</h1>

      <div className="mb-6">
        <p className="mb-3 text-sm font-semibold text-foreground">Available dates</p>
        {isLoadingDates ? (
          <Spinner className="h-5 w-5" />
        ) : availableDates.length === 0 ? (
          <p className="text-sm text-muted-foreground">No available dates found for this expert this month.</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {availableDates.map((date) => (
              <button
                key={date}
                type="button"
                onClick={() => setSelectedDate(date)}
                className={`rounded-xl border px-4 py-2 text-sm font-semibold transition-colors ${
                  selectedDate === date
                    ? "border-primary bg-primary text-white"
                    : "border-border text-foreground hover:bg-surface"
                }`}
              >
                {new Date(date).toLocaleDateString("en-GB", { day: "2-digit", month: "short" })}
              </button>
            ))}
          </div>
        )}
      </div>

      {selectedDate && (
        <div className="mb-8">
          <p className="mb-3 text-sm font-semibold text-foreground">Available times</p>
          {isLoadingTimes ? (
            <Spinner className="h-5 w-5" />
          ) : (timesResponse?.timeSlots.length ?? 0) === 0 ? (
            <p className="text-sm text-muted-foreground">No open slots on this date.</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {timesResponse!.timeSlots.map((slot) => (
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
                  {slot.startTime}
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
