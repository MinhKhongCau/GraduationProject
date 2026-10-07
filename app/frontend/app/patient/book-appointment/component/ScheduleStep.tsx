"use client";

import Image from "next/image";
import { useState, type ReactNode } from "react";
import { Check, ChevronDown, Clock } from "lucide-react";
import { WeekCalendar } from "./WeekCalendar";
import { PatientProfileSection } from "./PatientProfileSection";
import { Button, PageHeader, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { bookingApi, expertApi, specializationApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { AvailableTimeSlot, ExpertProfile, PatientRecord, Specialization } from "@/types";

/** Lists longer than these are cut short behind a "Show all" toggle. */
const VISIBLE_SPECIALIZATIONS = 8;
const VISIBLE_EXPERTS = 6;
const VISIBLE_TIMES = 12;

type SectionKey = "specialization" | "expert" | "schedule" | "profile";

export interface ScheduleStepProps {
  /** The page was opened for a specific expert, so start at their calendar. */
  isExpertPreselected: boolean;
  selectedSpecialization: Specialization | null;
  /** Passing null clears the filter so every expert is listed. */
  onSelectSpecialization: (specialization: Specialization | null) => void;
  selectedExpert: ExpertProfile | null;
  onSelectExpert: (expert: ExpertProfile) => void;
  selectedDate: string | null;
  onSelectDate: (date: string) => void;
  selectedSlot: AvailableTimeSlot | null;
  onSelectSlot: (slot: AvailableTimeSlot) => void;
  selectedRecord: PatientRecord | null;
  onSelectRecord: (record: PatientRecord) => void;
  onNext: () => void;
  /** True while the slot is being locked before moving to the confirm page. */
  isSubmitting: boolean;
}

/** Specialization, expert, date/time and patient profile picked together on one screen. */
export function ScheduleStep({
  isExpertPreselected,
  selectedSpecialization,
  onSelectSpecialization,
  selectedExpert,
  onSelectExpert,
  selectedDate,
  onSelectDate,
  selectedSlot,
  onSelectSlot,
  selectedRecord,
  onSelectRecord,
  onNext,
  isSubmitting,
}: ScheduleStepProps) {
  const specializationId = selectedSpecialization?.specId;
  const expertId = selectedExpert?.accountId ?? "";

  const [openSection, setOpenSection] = useState<SectionKey | null>(() => {
    if (!selectedExpert) return isExpertPreselected ? "schedule" : "specialization";
    if (!selectedSlot) return "schedule";
    if (!selectedRecord) return "profile";
    return null;
  });
  const [isCreatingRecord, setIsCreatingRecord] = useState(false);

  const { data: specializations = [], isLoading: isLoadingSpecializations } = useApiQuery({
    queryKey: QUERY_KEYS.specializations(),
    queryFn: () => specializationApi.getAllSpecializations(),
  });
  const activeSpecializations = specializations.filter((spec) => spec.isActive);

  const { data: experts = [], isLoading: isLoadingExperts } = useApiQuery({
    queryKey: QUERY_KEYS.experts({ specializationId }),
    queryFn: () => expertApi.getAllExperts(specializationId),
  });

  const { data: availableDates = [], isLoading: isLoadingDates } = useApiQuery({
    queryKey: QUERY_KEYS.availableDates(expertId),
    queryFn: () => bookingApi.getAvailableDates(expertId),
    enabled: !!expertId,
  });

  const { data: availableTimes = [], isLoading: isLoadingTimes } = useApiQuery({
    queryKey: QUERY_KEYS.availableTimes(expertId, selectedDate ?? ""),
    queryFn: () => bookingApi.getAvailableTimes(expertId, selectedDate!),
    enabled: !!expertId && !!selectedDate,
  });

  function toggleSection(key: SectionKey) {
    setOpenSection((current) => (current === key ? null : key));
  }

  function pickSpecialization(specialization: Specialization | null) {
    onSelectSpecialization(specialization);
    setOpenSection("expert");
  }

  function pickExpert(expert: ExpertProfile) {
    onSelectExpert(expert);
    setOpenSection("schedule");
  }

  function pickSlot(slot: AvailableTimeSlot) {
    onSelectSlot(slot);
    setOpenSection(selectedRecord ? null : "profile");
  }

  function pickRecord(record: PatientRecord) {
    onSelectRecord(record);
    setOpenSection(null);
  }

  return (
    <div className="flex h-full flex-col">
      <PageHeader
        title="Book an appointment"
        description="Pick a specialization, choose an expert, find a time and tell us who the session is for."
      />

      <div className="mb-8 flex flex-col gap-3">
        <Section
          index={1}
          title="Specialization"
          summary={selectedSpecialization?.name ?? "All specializations"}
          isDone={!!selectedSpecialization}
          isOpen={openSection === "specialization"}
          onToggle={() => toggleSection("specialization")}
        >
          {isLoadingSpecializations ? (
            <Spinner className="h-5 w-5" />
          ) : activeSpecializations.length === 0 ? (
            <p className="text-sm text-muted-foreground">No specializations are available right now.</p>
          ) : (
            <ShowMoreList items={activeSpecializations} limit={VISIBLE_SPECIALIZATIONS} className="flex flex-wrap gap-2">
              {(visible) => (
                <>
                  <Chip isSelected={!selectedSpecialization} onClick={() => pickSpecialization(null)}>
                    All
                  </Chip>
                  {visible.map((spec) => {
                    const isSelected = selectedSpecialization?.specId === spec.specId;
                    return (
                      <Chip
                        key={spec.specId}
                        isSelected={isSelected}
                        title={spec.description}
                        onClick={() => pickSpecialization(isSelected ? null : spec)}
                      >
                        {spec.name}
                      </Chip>
                    );
                  })}
                </>
              )}
            </ShowMoreList>
          )}
        </Section>

        <Section
          index={2}
          title="Expert"
          summary={selectedExpert?.fullName ?? "Not chosen yet"}
          isDone={!!selectedExpert}
          isOpen={openSection === "expert"}
          onToggle={() => toggleSection("expert")}
        >
          {isLoadingExperts ? (
            <Spinner className="h-5 w-5" />
          ) : experts.length === 0 ? (
            <p className="text-sm text-muted-foreground">No experts are available for this specialization yet.</p>
          ) : (
            <ShowMoreList items={experts} limit={VISIBLE_EXPERTS} className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {(visible) =>
                visible.map((expert) => {
                  const isSelected = selectedExpert?.accountId === expert.accountId;
                  return (
                    <button
                      key={expert.expertId}
                      type="button"
                      onClick={() => pickExpert(expert)}
                      aria-pressed={isSelected}
                      className={`relative flex items-center gap-3 rounded-xl border bg-background p-4 text-left transition-all ${
                        isSelected
                          ? "border-primary bg-primary-soft ring-1 ring-primary"
                          : "border-border-strong hover:border-primary/40 hover:shadow-card"
                      }`}
                    >
                      {isSelected && (
                        <div className="absolute right-3 top-3 text-primary">
                          <Check className="h-4 w-4 stroke-[3]" />
                        </div>
                      )}
                      <Image
                        src={expert.avatarUrl || "/images/auth-bg.png"}
                        alt={expert.fullName}
                        width={44}
                        height={44}
                        unoptimized
                        className="h-11 w-11 rounded-full object-cover"
                      />
                      <div className="min-w-0 pr-5">
                        <p className="truncate text-sm font-semibold text-foreground">{expert.fullName}</p>
                        <p className="line-clamp-1 text-xs text-muted-foreground">
                          {expert.specializations.map((spec) => spec.name).join(", ") || "General counseling"}
                        </p>
                      </div>
                    </button>
                  );
                })
              }
            </ShowMoreList>
          )}
        </Section>

        <Section
          index={3}
          title="Date & time"
          summary={selectedSlot ? formatSlot(selectedSlot) : "Not chosen yet"}
          isDone={!!selectedSlot}
          isOpen={openSection === "schedule"}
          onToggle={() => toggleSection("schedule")}
        >
          {!selectedExpert ? (
            <p className="text-sm text-muted-foreground">
              {isExpertPreselected ? "Loading the expert's calendar..." : "Choose an expert to see their available dates."}
            </p>
          ) : (
            <div className="flex flex-col gap-6">
              <WeekCalendar
                availableDates={availableDates}
                selectedDate={selectedDate}
                onSelectDate={onSelectDate}
                isLoading={isLoadingDates}
              />

              {selectedDate && (
                <div>
                  <p className="mb-3 text-sm font-semibold text-foreground">Available times</p>
                  {isLoadingTimes ? (
                    <Spinner className="h-5 w-5" />
                  ) : availableTimes.length === 0 ? (
                    <p className="text-sm text-muted-foreground">No open slots on this date.</p>
                  ) : (
                    <ShowMoreList items={availableTimes} limit={VISIBLE_TIMES} className="flex flex-wrap gap-2">
                      {(visible) =>
                        visible.map((slot) => (
                          <button
                            key={slot.slotId}
                            type="button"
                            onClick={() => pickSlot(slot)}
                            aria-pressed={selectedSlot?.slotId === slot.slotId}
                            className={`flex h-10 items-center gap-1.5 rounded-lg border px-4 text-sm font-semibold transition-colors ${
                              selectedSlot?.slotId === slot.slotId
                                ? "border-primary bg-primary text-white"
                                : "border-border-strong bg-background text-foreground hover:border-primary/40 hover:bg-surface"
                            }`}
                          >
                            <Clock className="h-3.5 w-3.5" />
                            {formatTime(slot.startTime)}
                          </button>
                        ))
                      }
                    </ShowMoreList>
                  )}
                </div>
              )}
            </div>
          )}
        </Section>

        <Section
          index={4}
          title="Patient profile"
          summary={selectedRecord ? selectedRecord.fullName || "Unnamed profile" : "Not chosen yet"}
          isDone={!!selectedRecord}
          isOpen={openSection === "profile"}
          onToggle={() => toggleSection("profile")}
        >
          <PatientProfileSection
            selectedRecord={selectedRecord}
            onSelectRecord={pickRecord}
            onCreatingChange={setIsCreatingRecord}
          />
        </Section>
      </div>

      <div className="mt-auto flex justify-end border-t border-border pt-6">
        <Button
          onClick={onNext}
          disabled={!selectedExpert || !selectedSlot || !selectedRecord || isCreatingRecord || isSubmitting}
        >
          {isSubmitting ? "Holding your slot..." : "Next"}
          {!isSubmitting && <Check className="h-4 w-4" />}
        </Button>
      </div>
    </div>
  );
}

function formatTime(value: number): string {
  return new Date(value).toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" });
}

function formatSlot(slot: AvailableTimeSlot): string {
  const date = new Date(slot.startTime).toLocaleDateString("en-GB", { weekday: "short", day: "2-digit", month: "short" });
  return `${date} · ${formatTime(slot.startTime)}`;
}

interface SectionProps {
  index: number;
  title: string;
  /** The current pick, shown in the header so a collapsed section still reads at a glance. */
  summary: string;
  isDone: boolean;
  isOpen: boolean;
  onToggle: () => void;
  children: ReactNode;
}

function Section({ index, title, summary, isDone, isOpen, onToggle, children }: SectionProps) {
  return (
    <section className="rounded-xl border border-border bg-background">
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={isOpen}
        className="flex w-full items-center gap-3 px-4 py-3 text-left"
      >
        <span
          className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
            isDone ? "bg-primary text-white" : "bg-primary-soft text-primary"
          }`}
        >
          {isDone ? <Check className="h-3.5 w-3.5 stroke-[3]" /> : index}
        </span>
        <span className="text-sm font-semibold text-foreground">{title}</span>
        <span className="ml-auto truncate text-sm text-muted-foreground">{summary}</span>
        <ChevronDown
          className={`h-4 w-4 shrink-0 text-muted-foreground transition-transform ${isOpen ? "rotate-180" : ""}`}
        />
      </button>
      {/* Kept mounted while collapsed so in-progress input (e.g. a new profile form) survives a toggle. */}
      <div hidden={!isOpen} className="border-t border-border p-4">
        {children}
      </div>
    </section>
  );
}

interface ShowMoreListProps<T> {
  items: T[];
  limit: number;
  className: string;
  children: (visible: T[]) => ReactNode;
}

function ShowMoreList<T>({ items, limit, className, children }: ShowMoreListProps<T>) {
  const [isExpanded, setIsExpanded] = useState(false);
  const hasMore = items.length > limit;
  const visible = hasMore && !isExpanded ? items.slice(0, limit) : items;

  return (
    <div>
      <div className={className}>{children(visible)}</div>
      {hasMore && (
        <button
          type="button"
          onClick={() => setIsExpanded((current) => !current)}
          className="mt-3 flex items-center gap-1 text-sm font-semibold text-primary hover:underline"
        >
          {isExpanded ? "Show less" : `Show all ${items.length}`}
          <ChevronDown className={`h-4 w-4 transition-transform ${isExpanded ? "rotate-180" : ""}`} />
        </button>
      )}
    </div>
  );
}

function Chip({
  isSelected,
  title,
  onClick,
  children,
}: {
  isSelected: boolean;
  title?: string;
  onClick: () => void;
  children: ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      title={title}
      aria-pressed={isSelected}
      className={`h-9 rounded-full border px-4 text-sm font-medium transition-colors ${
        isSelected
          ? "border-primary bg-primary text-white"
          : "border-border-strong bg-background text-foreground hover:border-primary/40 hover:bg-surface"
      }`}
    >
      {children}
    </button>
  );
}
