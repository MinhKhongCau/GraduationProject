"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import { ChevronLeft, ChevronRight, CalendarDays } from "lucide-react";
import { buttonClasses, Spinner } from "@/components/ui";

export interface WeekCalendarProps {
  /** "YYYY-MM-DD" dates the expert has open slots on. */
  availableDates: string[];
  selectedDate: string | null;
  onSelectDate: (date: string) => void;
  isLoading: boolean;
}

const DAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

function toDateKey(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function startOfWeek(date: Date): Date {
  const result = new Date(date);
  const dayIndex = (result.getDay() + 6) % 7; // 0 = Monday
  result.setDate(result.getDate() - dayIndex);
  result.setHours(0, 0, 0, 0);
  return result;
}

/** Week-at-a-glance date picker: patients browse the expert's available dates 7 days at a time. */
export function WeekCalendar({ availableDates, selectedDate, onSelectDate, isLoading }: WeekCalendarProps) {
  const availableSet = useMemo(() => new Set(availableDates), [availableDates]);
  const availableDateObjects = useMemo(
    () => availableDates.map((date) => new Date(`${date}T00:00:00`)),
    [availableDates]
  );

  const [weekStart, setWeekStart] = useState<Date>(() => startOfWeek(new Date()));
  const [pickerOpen, setPickerOpen] = useState(false);
  const seededRef = useRef(false);

  useEffect(() => {
    // One-time jump to the week of the first available date once dates load,
    // without clobbering the week the patient has since navigated to.
    if (!seededRef.current && availableDates.length > 0) {
      seededRef.current = true;
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setWeekStart(startOfWeek(new Date(`${availableDates[0]}T00:00:00`)));
    }
  }, [availableDates]);

  const weekDays = useMemo(
    () =>
      Array.from({ length: 7 }, (_, index) => {
        const date = new Date(weekStart);
        date.setDate(date.getDate() + index);
        return date;
      }),
    [weekStart]
  );

  function goToWeek(offsetDays: number) {
    setWeekStart((current) => {
      const next = new Date(current);
      next.setDate(next.getDate() + offsetDays);
      return next;
    });
  }

  function jumpTo(date: Date | null) {
    if (!date) return;
    setWeekStart(startOfWeek(date));
    onSelectDate(toDateKey(date));
    setPickerOpen(false);
  }

  return (
    <div>
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => goToWeek(-7)}
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-border-strong bg-background transition-colors hover:bg-surface"
            aria-label="Previous week"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <span className="text-sm font-semibold text-foreground">
            {weekDays[0].toLocaleDateString("en-GB", { day: "2-digit", month: "short" })} –{" "}
            {weekDays[6].toLocaleDateString("en-GB", { day: "2-digit", month: "short" })}
          </span>
          <button
            type="button"
            onClick={() => goToWeek(7)}
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-border-strong bg-background transition-colors hover:bg-surface"
            aria-label="Next week"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>

        <div className="relative">
          <button
            type="button"
            onClick={() => setPickerOpen((open) => !open)}
            aria-expanded={pickerOpen}
            className={buttonClasses("outline", "sm")}
          >
            <CalendarDays className="h-3.5 w-3.5" />
            Jump to date
          </button>
          {pickerOpen && (
            <div className="absolute right-0 z-20 mt-2">
              <DatePicker
                inline
                includeDates={availableDateObjects}
                selected={selectedDate ? new Date(`${selectedDate}T00:00:00`) : undefined}
                onChange={jumpTo}
              />
            </div>
          )}
        </div>
      </div>

      {isLoading ? (
        <Spinner className="h-5 w-5" />
      ) : availableDates.length === 0 ? (
        <p className="text-sm text-muted-foreground">No available dates found for this expert this month.</p>
      ) : (
        <div className="grid grid-cols-7 gap-2">
          {weekDays.map((date) => {
            const key = toDateKey(date);
            const isAvailable = availableSet.has(key);
            const isSelected = selectedDate === key;

            return (
              <button
                key={key}
                type="button"
                disabled={!isAvailable}
                onClick={() => onSelectDate(key)}
                aria-pressed={isSelected}
                className={`flex flex-col items-center gap-1 rounded-lg border py-3 text-xs font-semibold transition-colors ${
                  isSelected
                    ? "border-primary bg-primary text-white"
                    : isAvailable
                      ? "border-border-strong bg-background text-foreground hover:border-primary/40 hover:bg-surface"
                      : "cursor-not-allowed border-border bg-surface text-muted-foreground/50"
                }`}
              >
                <span>{DAY_LABELS[(date.getDay() + 6) % 7]}</span>
                <span className="text-base font-bold">{date.getDate()}</span>
                {isAvailable && !isSelected && <span className="h-1 w-1 rounded-full bg-primary" />}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
