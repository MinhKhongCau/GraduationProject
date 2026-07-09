"use client";

import { useState } from "react";
import { Modal, Button } from "@/components/ui";
import type { AppointmentWithExpert } from "@/hooks";

export interface CancelBookingDialogProps {
  appointment: AppointmentWithExpert | null;
  onOpenChange: (open: boolean) => void;
  onConfirm: (reason: string) => void;
  isSubmitting: boolean;
}

export function CancelBookingDialog({ appointment, onOpenChange, onConfirm, isSubmitting }: CancelBookingDialogProps) {
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
      open={!!appointment}
      onOpenChange={handleOpenChange}
      title="Cancel this booking?"
      description={
        appointment ? `Your session with ${appointment.expertName ?? "the expert"} will be canceled.` : undefined
      }
    >
      <div className="space-y-4">
        <div>
          <label className="mb-1 block text-xs font-bold text-foreground">Reason for cancellation</label>
          <textarea
            rows={3}
            value={reason}
            onChange={(event) => setReason(event.target.value)}
            disabled={isSubmitting}
            className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
          />
          {touched && !trimmedReason && <p className="mt-1 text-xs text-danger">A reason is required.</p>}
        </div>

        <div className="flex justify-end gap-3">
          <Button variant="outline" onClick={() => handleOpenChange(false)} disabled={isSubmitting}>
            Keep booking
          </Button>
          <Button variant="danger" onClick={handleConfirm} disabled={isSubmitting}>
            {isSubmitting ? "Canceling..." : "Cancel booking"}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
