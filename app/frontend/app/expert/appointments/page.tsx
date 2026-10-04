"use client";

import { AppointmentListItem } from "./component/AppointmentListItem";
import { Card, PageHeader, Spinner } from "@/components/ui";
import { useExpertAppointments } from "@/hooks";

export default function ExpertAppointmentsPage() {
  const { data: appointments = [], isLoading } = useExpertAppointments();

  return (
    <div className="mx-auto max-w-4xl">
      <PageHeader title="My Appointments" />

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : appointments.length === 0 ? (
        <Card className="p-8 text-center text-sm text-muted-foreground">You don&apos;t have any appointments yet.</Card>
      ) : (
        <div className="space-y-3">
          {appointments.map((appointment) => (
            <AppointmentListItem key={appointment.appointmentId} appointment={appointment} />
          ))}
        </div>
      )}
    </div>
  );
}
