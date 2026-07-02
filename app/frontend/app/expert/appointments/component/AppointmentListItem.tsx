import { CalendarClock } from "lucide-react";
import { Card } from "@/components/ui";
import type { Appointment } from "@/types";

const STATUS_STYLES: Record<Appointment["status"], string> = {
  PENDING: "bg-warning-soft text-warning",
  LOCKED: "bg-warning-soft text-warning",
  CONFIRMED: "bg-success-soft text-success",
  CANCELED: "bg-danger-soft text-danger",
  COMPLETED: "bg-surface text-muted-foreground",
};

export function AppointmentListItem({ appointment }: { appointment: Appointment }) {
  return (
    <Card className="flex items-center justify-between p-5">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary-soft-text">
          <CalendarClock className="h-5 w-5" />
        </div>
        <div>
          <p className="text-sm font-bold text-foreground">{appointment.topic || "Consultation"}</p>
          <p className="text-xs text-muted-foreground">
            Booked {new Date(appointment.createdAt).toLocaleString("en-GB")}
          </p>
        </div>
      </div>
      <span className={`rounded-md px-2 py-1 text-[10px] font-bold uppercase ${STATUS_STYLES[appointment.status]}`}>
        {appointment.status}
      </span>
    </Card>
  );
}
