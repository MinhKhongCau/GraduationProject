"use client";

import { useState } from "react";
import { Modal, Button, Label, Textarea, FieldError } from "@/components/ui";
import type { Appointment } from "@/types";

export interface CancelAppointmentsDialogProps<T extends Appointment> {
  /** Appointments to cancel; empty closes the dialog. */
  appointments: T[];
  /** Name of the other party shown in the single-appointment description (expert for patients, patient for experts). */
  counterpartName: (appointment: T) => string | undefined;
  /** Used when `counterpartName` returns nothing, e.g. "the expert". */
  counterpartFallback: string;
  onOpenChange: (open: boolean) => void;
  onConfirm: (reason: string) => void;
  isSubmitting: boolean;
}

export function CancelAppointmentsDialog<T extends Appointment>({
  appointments,
  counterpartName,
  counterpartFallback,
  onOpenChange,
  onConfirm,
  isSubmitting,
}: CancelAppointmentsDialogProps<T>) {
  const [reason, setReason] = useState("");
  const [touched, setTouched] = useState(false);
  const trimmedReason = reason.trim();

  function handleOpenChange(open: boolean) {
    if (!open) {
      setReason("");
      setTouched(false);
    }
    onOpenChange(open);
  }

  function handleConfirm() {
    if (!trimmedReason) {
      setTouched(true);
      return;
    }
    onConfirm(trimmedReason);
  }

  return (
    <Modal
      open={appointments.length > 0}
      onOpenChange={handleOpenChange}
      title={appointments.length > 1 ? `Cancel ${appointments.length} bookings?` : "Cancel this booking?"}
      description={
        appointments.length > 1
          ? `${appointments.length} sessions will be canceled with the same reason.`
          : appointments[0]
            ? `Your session with ${counterpartName(appointments[0]) ?? counterpartFallback} will be canceled.`
            : undefined
      }
    >
      <div className="space-y-4">
        <div>
          <Label htmlFor="cancel-reason">Reason for cancellation</Label>
          <Textarea
            id="cancel-reason"
            rows={3}
            invalid={touched && !trimmedReason}
            value={reason}
            onChange={(event) => setReason(event.target.value)}
            disabled={isSubmitting}
          />
          <FieldError>{touched && !trimmedReason ? "A reason is required." : undefined}</FieldError>
        </div>

        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <Button variant="outline" onClick={() => handleOpenChange(false)} disabled={isSubmitting}>
            Keep booking
          </Button>
          <Button variant="danger" onClick={handleConfirm} disabled={isSubmitting}>
            {isSubmitting ? "Canceling..." : appointments.length > 1 ? "Cancel bookings" : "Cancel booking"}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
