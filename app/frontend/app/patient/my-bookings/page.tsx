"use client";

import { useState } from "react";
import { BookingListItem } from "./component/BookingListItem";
import { CancelBookingDialog } from "./component/CancelBookingDialog";
import { Spinner } from "@/components/ui";
import { useMyBookings, useCancelBooking, usePaymentRedirect } from "@/hooks";
import { useApiMutation } from "@/hooks";
import { paymentApi } from "@/api";
import { CHANNELING_FEE } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type { AppointmentWithExpert } from "@/hooks";

export default function MyBookingsPage() {
  const [cancelingAppointment, setCancelingAppointment] = useState<AppointmentWithExpert | null>(null);
  const [payingAppointmentId, setPayingAppointmentId] = useState<string | null>(null);
  const { showSuccess, showError } = useErrorContext();

  const { data: appointments = [], isLoading, refetch } = useMyBookings();
  const cancelMutation = useCancelBooking();
  const { isPaymentBrowserOpen, openPayment } = usePaymentRedirect({
    onBrowserFinished: () => {
      setPayingAppointmentId(null);
      void refetch();
      showError({
        message: "Payment browser closed. Your booking may still be pending payment; please check its status before trying again.",
      });
    },
  });

  const payMutation = useApiMutation({
    mutationFn: (appointment: AppointmentWithExpert) =>
      paymentApi.createOrder({
        expertId: appointment.expertId,
        amount: CHANNELING_FEE,
        gateway: "VNPAY",
        appointmentId: appointment.appointmentId,
      }),
    onSuccess: async (order) => {
      const opened = await openPayment(order.paymentUrl);
      if (!opened) {
        setPayingAppointmentId(null);
        showError({ message: "Unable to open the payment page. Please try again." });
      }
    },
  });

  function handlePayNow(appointment: AppointmentWithExpert) {
    setPayingAppointmentId(appointment.appointmentId);
    payMutation.mutate(appointment);
  }

  function handleConfirmCancel(reason: string) {
    if (!cancelingAppointment) return;
    cancelMutation.mutate(
      { appointmentId: cancelingAppointment.appointmentId, reason },
      {
        onSuccess: () => {
          showSuccess("Booking canceled.");
          setCancelingAppointment(null);
        },
      }
    );
  }

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
              onPayNow={handlePayNow}
              isPaying={(payMutation.isPending || isPaymentBrowserOpen) && payingAppointmentId === appointment.appointmentId}
            />
          ))}
        </div>
      )}

      <CancelBookingDialog
        appointment={cancelingAppointment}
        onOpenChange={(open) => !open && setCancelingAppointment(null)}
        onConfirm={handleConfirmCancel}
        isSubmitting={cancelMutation.isPending}
      />
    </div>
  );
}
