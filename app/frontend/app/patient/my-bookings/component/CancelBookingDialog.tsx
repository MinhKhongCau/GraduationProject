import { Modal, Button } from "@/components/ui";
import type { Appointment } from "@/types";

export interface CancelBookingDialogProps {
  appointment: Appointment | null;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isSubmitting: boolean;
}

export function CancelBookingDialog({ appointment, onOpenChange, onConfirm, isSubmitting }: CancelBookingDialogProps) {
  return (
    <Modal
      open={!!appointment}
      onOpenChange={onOpenChange}
      title="Cancel this booking?"
      description={appointment ? `Your session with ${appointment.expertName ?? "the expert"} will be canceled.` : undefined}
    >
      <div className="flex justify-end gap-3">
        <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
          Keep booking
        </Button>
        <Button variant="danger" onClick={onConfirm} disabled={isSubmitting}>
          {isSubmitting ? "Canceling..." : "Cancel booking"}
        </Button>
      </div>
    </Modal>
  );
}
