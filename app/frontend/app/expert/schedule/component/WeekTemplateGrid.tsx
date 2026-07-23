"use client";

import { useState } from "react";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import { CalendarRange, ChevronDown } from "lucide-react";
import { Button, Card } from "@/components/ui";
import type { TimeTemplate, Availability, DayOfWeek } from "@/types";

const DAYS: { value: DayOfWeek; short: string }[] = [
  { value: 1, short: "Mon" },
  { value: 2, short: "Tue" },
  { value: 3, short: "Wed" },
  { value: 4, short: "Thu" },
  { value: 5, short: "Fri" },
  { value: 6, short: "Sat" },
  { value: 7, short: "Sun" },
];

export interface DayTemplateConfig {
  enabled: boolean;
  templateId: string;
  effectiveFrom: Date;
  effectiveUntil: Date | null;
}

export type WeekTemplateState = Record<DayOfWeek, DayTemplateConfig>;

export interface WeekTemplateGridProps {
  templates: TimeTemplate[];
  availabilities: Availability[];
  onSave: (state: WeekTemplateState) => void;
  isSubmitting: boolean;
}

function buildInitialState(templates: TimeTemplate[], availabilities: Availability[]): WeekTemplateState {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const defaultTemplateId = templates.find((t) => t.isActive)?.templateId ?? templates[0]?.templateId ?? "";

  const state = {} as WeekTemplateState;
  for (const day of DAYS) {
    const existing = availabilities.find((a) => a.dayOfWeek === day.value);
    state[day.value] = existing
      ? {
          enabled: existing.isEnabled,
          templateId: existing.templateId,
          effectiveFrom: new Date(existing.effectiveFrom),
          effectiveUntil: existing.effectiveUntil ? new Date(existing.effectiveUntil) : null,
        }
      : { enabled: false, templateId: defaultTemplateId, effectiveFrom: today, effectiveUntil: null };
  }
  return state;
}

/** 7-day weekly template builder: for each weekday, expert enables it and picks a shift template. */
export function WeekTemplateGrid({ templates, availabilities, onSave, isSubmitting }: WeekTemplateGridProps) {
  const [state, setState] = useState<WeekTemplateState>(() => buildInitialState(templates, availabilities));
  const [expandedDay, setExpandedDay] = useState<DayOfWeek | null>(null);

  function updateDay(day: DayOfWeek, changes: Partial<DayTemplateConfig>) {
    setState((current) => ({ ...current, [day]: { ...current[day], ...changes } }));
  }

  const activeTemplates = templates.filter((template) => template.isActive);

  return (
    <Card className="p-6">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-7">
        {DAYS.map((day) => {
          const config = state[day.value];
          const template = activeTemplates.find((t) => t.templateId === config.templateId);
          const isExpanded = expandedDay === day.value;

          return (
            <div
              key={day.value}
              className={`flex flex-col gap-2 rounded-xl border p-3 transition-colors ${
                config.enabled ? "border-primary bg-primary-soft/40" : "border-border"
              }`}
            >
              <label className="flex items-center justify-between gap-2 text-sm font-bold text-foreground">
                {day.short}
                <input
                  type="checkbox"
                  checked={config.enabled}
                  onChange={(event) => updateDay(day.value, { enabled: event.target.checked })}
                />
              </label>

              {config.enabled && (
                <>
                  <select
                    value={config.templateId}
                    onChange={(event) => updateDay(day.value, { templateId: event.target.value })}
                    className="rounded-lg border border-border px-2 py-1.5 text-xs outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
                  >
                    {activeTemplates.length === 0 && <option value="">No templates yet</option>}
                    {activeTemplates.map((t) => (
                      <option key={t.templateId} value={t.templateId}>
                        {t.shiftName}
                      </option>
                    ))}
                  </select>

                  {template && (
                    <p className="text-[11px] text-muted-foreground">
                      {template.startTime}–{template.endTime} · {template.slotDurationMinutes}min slots
                    </p>
                  )}

                  <button
                    type="button"
                    onClick={() => setExpandedDay(isExpanded ? null : day.value)}
                    className="flex items-center gap-1 text-[11px] font-semibold text-primary hover:underline"
                  >
                    <CalendarRange className="h-3 w-3" />
                    Effective dates
                    <ChevronDown className={`h-3 w-3 transition-transform ${isExpanded ? "rotate-180" : ""}`} />
                  </button>

                  {isExpanded && (
                    <div className="space-y-1.5 rounded-lg bg-surface p-2">
                      <div>
                        <p className="mb-0.5 text-[10px] font-semibold text-muted-foreground">From</p>
                        <DatePicker
                          selected={config.effectiveFrom}
                          onChange={(date: Date | null) => date && updateDay(day.value, { effectiveFrom: date })}
                          dateFormat="dd/MM/yyyy"
                          className="w-full rounded-lg border border-border px-2 py-1 text-xs outline-none focus:border-primary"
                        />
                      </div>
                      <div>
                        <p className="mb-0.5 text-[10px] font-semibold text-muted-foreground">Until (optional)</p>
                        <DatePicker
                          selected={config.effectiveUntil}
                          onChange={(date: Date | null) => updateDay(day.value, { effectiveUntil: date })}
                          dateFormat="dd/MM/yyyy"
                          isClearable
                          minDate={config.effectiveFrom}
                          placeholderText="No end date"
                          className="w-full rounded-lg border border-border px-2 py-1 text-xs outline-none focus:border-primary"
                        />
                      </div>
                    </div>
                  )}
                </>
              )}
            </div>
          );
        })}
      </div>

      <Button className="mt-6" onClick={() => onSave(state)} disabled={isSubmitting}>
        {isSubmitting ? "Saving..." : "Save weekly template"}
      </Button>
    </Card>
  );
}
