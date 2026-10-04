import Link from "next/link";
import { CalendarClock, ArrowRight } from "lucide-react";
import { Card, buttonClasses } from "@/components/ui";
import { ROUTES } from "@/constants";
import type { AppointmentWithExpert } from "@/hooks";

export function UpcomingAppointmentCard({ appointment }: { appointment: AppointmentWithExpert | null }) {
  if (!appointment) {
    return (
      <Card className="flex flex-col items-start gap-3 p-6">
        <CalendarClock className="h-8 w-8 text-muted-foreground" />
        <div>
          <h3 className="font-semibold text-foreground">No upcoming appointments</h3>
          <p className="text-sm text-muted-foreground">Book a session with an expert to get started.</p>
        </div>
        <Link href={ROUTES.PATIENT.BOOK_APPOINTMENT} className={buttonClasses("primary", "sm")}>
          Book an appointment
        </Link>
      </Card>
    );
  }

  return (
    <Card className="p-6">
      <p className="mb-1 text-xs font-semibold uppercase tracking-wider text-primary">Upcoming appointment</p>
      <h3 className="mb-1 text-lg font-semibold text-foreground">{appointment.expertName ?? "Expert"}</h3>
      <p className="mb-4 text-sm text-muted-foreground">
        {new Date(appointment.createdAt).toLocaleString("en-GB")}
      </p>
      <Link href={ROUTES.PATIENT.MY_BOOKINGS} className="inline-flex items-center gap-1 text-sm font-semibold text-primary hover:underline">
        View my bookings <ArrowRight className="h-3.5 w-3.5" />
      </Link>
    </Card>
  );
}
