"use client";

import { useState } from "react";
import { BookingListItem } from "./component/BookingListItem";
import { CancelBookingDialog } from "./component/CancelBookingDialog";
import { Spinner } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useQueryClient } from "@tanstack/react-query";
import { useErrorContext } from "@/context/ErrorContext";
import type { Appointment } from "@/types";

export default function MyBookingsPage() {
  const [cancelingAppointment, setCancelingAppointment] = useState<Appointment | null>(null);
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();

  const { data: appointments = [], isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.bookingHistory(),
    queryFn: () => bookingApi.getBookingHistory(),
  });

  const cancelMutation = useApiMutation({
    mutationFn: (appointmentId: string) => bookingApi.cancelAppointment(appointmentId),
    onSuccess: () => {
      showSuccess("Booking canceled.");
      setCancelingAppointment(null);
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.bookingHistory() });
    },
  });

  return (
    <div className="mx-auto max-w-4xl">
      <h1 className="mb-6 text-2xl font-bold text-foreground">My Bookings</h1>

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : appointments.length === 0 ? (
        <p className="text-sm text-muted-foreground">You don&apos;t have any bookings yet.</p>
      ) : (
        <div className="space-y-3">
          {appointments.map((appointment) => (
            <BookingListItem
              key={appointment.appointmentId}
              appointment={appointment}
              onCancel={setCancelingAppointment}
            />
          ))}
        </div>
      )}

      <CancelBookingDialog
        appointment={cancelingAppointment}
        onOpenChange={(open) => !open && setCancelingAppointment(null)}
        onConfirm={() => cancelingAppointment && cancelMutation.mutate(cancelingAppointment.appointmentId)}
        isSubmitting={cancelMutation.isPending}
      />
    </div>
  );
}
