import { CalendarClock } from "lucide-react";
import { Card, Button } from "@/components/ui";
import type { Appointment } from "@/types";

const STATUS_STYLES: Record<Appointment["status"], string> = {
  PENDING: "bg-warning-soft text-warning",
  LOCKED: "bg-warning-soft text-warning",
  CONFIRMED: "bg-success-soft text-success",
  CANCELED: "bg-danger-soft text-danger",
  COMPLETED: "bg-surface text-muted-foreground",
};

export interface BookingListItemProps {
  appointment: Appointment;
  onCancel: (appointment: Appointment) => void;
}

export function BookingListItem({ appointment, onCancel }: BookingListItemProps) {
  const canCancel = appointment.status === "CONFIRMED" || appointment.status === "PENDING";

  return (
    <Card className="flex flex-col gap-3 p-5 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-start gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary-soft-text">
          <CalendarClock className="h-5 w-5" />
        </div>
        <div>
          <p className="text-sm font-bold text-foreground">{appointment.expertName ?? "Expert"}</p>
          <p className="text-xs text-muted-foreground">{appointment.topic}</p>
          <p className="mt-1 text-xs text-muted-foreground">
            Booked {new Date(appointment.createdAt).toLocaleString("en-GB")}
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <span className={`rounded-md px-2 py-1 text-[10px] font-bold uppercase ${STATUS_STYLES[appointment.status]}`}>
          {appointment.status}
        </span>
        {canCancel && (
          <Button size="sm" variant="outline" onClick={() => onCancel(appointment)}>
            Cancel
          </Button>
        )}
      </div>
    </Card>
  );
}
