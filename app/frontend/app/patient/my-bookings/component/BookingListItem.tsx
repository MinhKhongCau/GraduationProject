import { CalendarClock, Video } from "lucide-react";
import { Card, Button } from "@/components/ui";
import type { AppointmentStatus } from "@/types";
import type { AppointmentWithExpert } from "@/hooks";

const STATUS_STYLES: Record<AppointmentStatus, string> = {
  PENDING_PAYMENT: "bg-warning-soft text-warning",
  CONFIRMED: "bg-success-soft text-success",
  CANCELLED: "bg-danger-soft text-danger",
};

const STATUS_TEXT: Record<AppointmentStatus, string> = {
  PENDING_PAYMENT: "Pending payment",
  CONFIRMED: "Confirmed",
  CANCELLED: "Cancelled",
};

export interface BookingListItemProps {
  appointment: AppointmentWithExpert;
  onCancel: (appointment: AppointmentWithExpert) => void;
  onPayNow: (appointment: AppointmentWithExpert) => void;
  isPaying: boolean;
}

export function BookingListItem({ appointment, onCancel, onPayNow, isPaying }: BookingListItemProps) {
  const canCancel = appointment.statusLabel === "CONFIRMED" || appointment.statusLabel === "PENDING_PAYMENT";
  const canPay = appointment.statusLabel === "PENDING_PAYMENT";

  return (
    <Card className="flex flex-col gap-3 p-5 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-start gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary-soft-text">
          <CalendarClock className="h-5 w-5" />
        </div>
        <div>
          <p className="text-sm font-bold text-foreground">{appointment.expertName ?? "Expert"}</p>
          <p className="mt-1 text-xs text-muted-foreground">
            Booked {new Date(appointment.createdAt).toLocaleString("en-GB")}
          </p>
          {appointment.statusLabel === "CANCELLED" && appointment.cancellationReason && (
            <p className="mt-1 text-xs text-muted-foreground">Reason: {appointment.cancellationReason}</p>
          )}
          {appointment.statusLabel === "CONFIRMED" && appointment.meetingLink && (
            <a
              href={appointment.meetingLink}
              target="_blank"
              rel="noreferrer"
              className="mt-1 inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline"
            >
              <Video className="h-3 w-3" />
              Join meeting
            </a>
          )}
        </div>
      </div>

      <div className="flex items-center gap-3">
        <span
          className={`rounded-md px-2 py-1 text-[10px] font-bold uppercase ${STATUS_STYLES[appointment.statusLabel]}`}
        >
          {STATUS_TEXT[appointment.statusLabel]}
        </span>
        {canPay && (
          <Button size="sm" onClick={() => onPayNow(appointment)} disabled={isPaying}>
            {isPaying ? "Redirecting..." : "Pay now"}
          </Button>
        )}
        {canCancel && (
          <Button size="sm" variant="outline" onClick={() => onCancel(appointment)}>
            Cancel
          </Button>
        )}
      </div>
    </Card>
  );
}
