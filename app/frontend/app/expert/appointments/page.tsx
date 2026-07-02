"use client";

import { AppointmentListItem } from "./component/AppointmentListItem";
import { Spinner } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";

export default function ExpertAppointmentsPage() {
  const { user } = useAuthContext();

  const { data: appointments = [], isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.bookingHistory(),
    queryFn: () => bookingApi.getBookingHistory(),
  });

  const myAppointments = appointments.filter((appointment) => appointment.expertId === user?.id);

  return (
    <div className="mx-auto max-w-4xl">
      <h1 className="mb-6 text-2xl font-bold text-foreground">My Appointments</h1>

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : myAppointments.length === 0 ? (
        <p className="text-sm text-muted-foreground">You don&apos;t have any appointments yet.</p>
      ) : (
        <div className="space-y-3">
          {myAppointments.map((appointment) => (
            <AppointmentListItem key={appointment.appointmentId} appointment={appointment} />
          ))}
        </div>
      )}
    </div>
  );
}
