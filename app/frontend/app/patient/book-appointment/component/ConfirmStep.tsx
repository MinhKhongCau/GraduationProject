"use client";

import { useEffect, useState, type ReactNode } from "react";
import Image from "next/image";
import { AlertTriangle, CalendarCheck, Timer } from "lucide-react";
import { Button, Card, PageHeader, Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { bookingApi } from "@/api";
import { normalizeError } from "@/api/http/errorNormalizer";
import { QUERY_KEYS } from "@/constants";
import type { BookingConfirmationParams, PatientRecordRelationship } from "@/types";
import { RELATIONSHIP_LABELS } from "./PatientProfileSection";

export interface ConfirmStepProps {
  params: BookingConfirmationParams;
  onBack: () => void;
  /** Hold expired or lost — send the patient back to pick a slot. */
  onPickAnotherSlot: () => void;
  onConfirm: () => void;
  isSubmitting: boolean;
}

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-right text-sm font-semibold text-foreground">{children}</span>
    </div>
  );
}

function formatCountdown(ms: number): string {
  const totalSeconds = Math.max(0, Math.floor(ms / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

/** Final review: slot, expert and patient profile as resolved by booking-service (via profile-service gRPC). */
export function ConfirmStep({ params, onBack, onPickAnotherSlot, onConfirm, isSubmitting }: ConfirmStepProps) {
  const { data: confirmation, isLoading, error, refetch, isFetching } = useApiQuery({
    queryKey: QUERY_KEYS.bookingConfirmation({ ...params }),
    queryFn: () => bookingApi.getBookingConfirmation(params),
    retry: false,
    staleTime: 0,
  });
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, []);

  if (isLoading) {
    return <Spinner className="h-6 w-6" />;
  }

  if (error || !confirmation) {
    const statusCode = error ? normalizeError(error).statusCode : undefined;
    const holdLost = statusCode === 409;
    return (
      <div className="flex h-full flex-col items-center justify-center text-center">
        <AlertTriangle className="mb-4 h-12 w-12 text-danger" />
        <h1 className="mb-2 text-2xl font-bold text-foreground">
          {holdLost ? "Your slot hold has ended" : "Couldn't load your booking details"}
        </h1>
        <p className="mb-8 max-w-md text-sm text-muted-foreground">
          {holdLost
            ? "This time slot is no longer reserved for you. Please pick a time again."
            : "Something went wrong while loading the booking details. Please try again."}
        </p>
        <div className="flex gap-3">
          <Button variant="outline" onClick={holdLost ? onPickAnotherSlot : onBack}>
            {holdLost ? "Pick another slot" : "Back"}
          </Button>
          {!holdLost && (
            <Button onClick={() => refetch()} disabled={isFetching}>
              {isFetching ? "Retrying..." : "Retry"}
            </Button>
          )}
        </div>
      </div>
    );
  }

  const { slot, expert, patientRecord, specialization } = confirmation;
  const remainingMs = slot.lockedExpiresAt - now;
  const holdExpired = remainingMs <= 0;
  const start = new Date(slot.startTime);
  const end = new Date(slot.endTime);
  const dateLabel = start.toLocaleDateString("en-GB", { weekday: "short", day: "2-digit", month: "short", year: "numeric" });
  const timeLabel = `${start.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" })} – ${end.toLocaleTimeString("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
  })}`;
  const relationshipLabel =
    RELATIONSHIP_LABELS[patientRecord.relationship as PatientRecordRelationship] ?? patientRecord.relationship;

  return (
    <div className="flex h-full flex-col">
      <PageHeader title="Confirm your booking" />

      <div
        className={`mb-6 flex max-w-xl items-center gap-2 rounded-lg border px-4 py-3 text-sm ${
          holdExpired ? "border-danger/30 bg-danger-soft text-danger" : "border-primary/20 bg-primary-soft text-primary-soft-text"
        }`}
        role="status"
      >
        <Timer className="h-4 w-4 shrink-0" />
        {holdExpired ? (
          <span>Your hold on this slot has expired. Pick another time to continue.</span>
        ) : (
          <span>
            Slot held for you for <span className="font-bold tabular-nums">{formatCountdown(remainingMs)}</span>
          </span>
        )}
      </div>

      <div className="mb-6 grid max-w-3xl grid-cols-1 gap-4 md:grid-cols-2">
        <Card className="space-y-4 p-6">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Appointment</p>
          <div className="flex items-center gap-3">
            <Image
              src={expert.avatarUrl || "/images/auth-bg.png"}
              alt={expert.fullName}
              width={44}
              height={44}
              unoptimized
              className="h-11 w-11 rounded-full object-cover"
            />
            <div>
              <p className="text-sm font-semibold text-foreground">{expert.fullName}</p>
              <p className="text-xs text-muted-foreground">
                {expert.specializations.map((spec) => spec.name).join(", ") || "General counseling"}
              </p>
            </div>
          </div>
          {specialization && <Row label="Specialization">{specialization.name}</Row>}
          <Row label="Date">
            <span className="flex items-center gap-1.5">
              <CalendarCheck className="h-4 w-4 text-primary" />
              {dateLabel}
            </span>
          </Row>
          <Row label="Time">{timeLabel}</Row>
          <div className="flex items-center justify-between border-t border-border pt-4">
            <span className="text-sm text-muted-foreground">Channeling fee</span>
            <span className="text-lg font-extrabold text-foreground">{slot.price.toLocaleString("vi-VN")} đ</span>
          </div>
        </Card>

        <Card className="space-y-4 p-6">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Patient</p>
          <Row label="Full name">{patientRecord.fullName}</Row>
          <Row label="Relationship">{relationshipLabel}</Row>
          {patientRecord.dateOfBirth && <Row label="Date of birth">{patientRecord.dateOfBirth}</Row>}
          {patientRecord.gender && <Row label="Gender">{patientRecord.gender}</Row>}
          {patientRecord.phoneNumber && <Row label="Phone">{patientRecord.phoneNumber}</Row>}
          {patientRecord.email && <Row label="Email">{patientRecord.email}</Row>}
          {patientRecord.address && <Row label="Address">{patientRecord.address}</Row>}
        </Card>
      </div>

      <p className="mb-6 max-w-xl text-xs text-muted-foreground">
        Confirming creates your booking and takes you straight to VNPay to complete payment.
      </p>

      <div className="mt-auto flex justify-between border-t border-border pt-6">
        <Button variant="outline" onClick={onBack} disabled={isSubmitting}>
          Back
        </Button>
        {holdExpired ? (
          <Button onClick={onPickAnotherSlot}>Pick another slot</Button>
        ) : (
          <Button onClick={onConfirm} disabled={isSubmitting}>
            {isSubmitting ? "Booking..." : "Confirm booking"}
          </Button>
        )}
      </div>
    </div>
  );
}
