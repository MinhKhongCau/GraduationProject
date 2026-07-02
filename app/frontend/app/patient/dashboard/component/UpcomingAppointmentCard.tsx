import Link from "next/link";
import { CalendarClock, ArrowRight } from "lucide-react";
import { Card, Button } from "@/components/ui";
import { ROUTES } from "@/constants";
import type { Appointment } from "@/types";

export function UpcomingAppointmentCard({ appointment }: { appointment: Appointment | null }) {
  if (!appointment) {
    return (
      <Card className="flex flex-col items-start gap-3 p-6">
        <CalendarClock className="h-8 w-8 text-muted-foreground" />
        <div>
          <h3 className="font-bold text-foreground">No upcoming appointments</h3>
          <p className="text-sm text-muted-foreground">Book a session with an expert to get started.</p>
        </div>
        <Link href={ROUTES.PATIENT.BOOK_APPOINTMENT}>
          <Button size="sm">Book an appointment</Button>
        </Link>
      </Card>
    );
  }

  return (
    <Card className="p-6">
      <p className="mb-1 text-xs font-bold uppercase tracking-wider text-primary">Upcoming appointment</p>
      <h3 className="mb-1 text-lg font-bold text-foreground">{appointment.expertName}</h3>
      <p className="mb-4 text-sm text-muted-foreground">{appointment.topic}</p>
      <Link href={ROUTES.PATIENT.MY_BOOKINGS} className="inline-flex items-center gap-1 text-sm font-semibold text-primary hover:underline">
        View my bookings <ArrowRight className="h-3.5 w-3.5" />
      </Link>
    </Card>
  );
}
