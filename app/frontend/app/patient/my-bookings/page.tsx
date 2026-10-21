"use client";

import { useState } from "react";
import { BookingTableRow } from "./component/BookingTableRow";
import {
  AppointmentListHeader,
  AppointmentListSummary,
  AppointmentTable,
  CancelAppointmentsDialog,
  DateSortFilters,
  STATUSES,
  StatusFilterLinks,
  useAppointmentListState,
} from "@/components/appointment";
import { useMyBookingsPage, useMyBookingCounts, useCancelBooking, usePaymentRedirect } from "@/hooks";
import { useApiMutation } from "@/hooks";
import { paymentApi } from "@/api";
import { CHANNELING_FEE } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type { AppointmentWithExpert } from "@/hooks";

const expertName = (a: AppointmentWithExpert) => a.expertName;

export default function MyBookingsPage() {
  const list = useAppointmentListState();
  const [cancelingAppointments, setCancelingAppointments] = useState<AppointmentWithExpert[]>([]);
  const [payingAppointmentId, setPayingAppointmentId] = useState<string | null>(null);
  const [isCanceling, setIsCanceling] = useState(false);
  const { showSuccess, showError } = useErrorContext();

  const { data: bookingPage, isLoading, isFetching, refetch } = useMyBookingsPage(list.params);
  const filterCounts = useMyBookingCounts(STATUSES, list.dateRange);
  const cancelMutation = useCancelBooking();
  const { isPaymentBrowserOpen, openPayment } = usePaymentRedirect({
    onBrowserFinished: () => {
      setPayingAppointmentId(null);
      void refetch();
      showError({
        message:
          "Payment browser closed. Your booking may still be pending payment; please check its status before trying again.",
      });
    },
  });

  const payMutation = useApiMutation({
    mutationFn: (appointment: AppointmentWithExpert) =>
      paymentApi.createOrder({
        expertId: appointment.expertId,
        amount: appointment.price ?? CHANNELING_FEE,
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

  const visibleAppointments = bookingPage?.items ?? [];
  const totalPages = bookingPage?.totalPages ?? 0;

  // Cancelling the last rows of the last page (under a status filter) can leave `page` past the end.
  if (bookingPage && !isFetching && list.page > 0 && list.page >= totalPages) {
    list.setPage(Math.max(totalPages - 1, 0));
  }

  function handlePayNow(appointment: AppointmentWithExpert) {
    setPayingAppointmentId(appointment.appointmentId);
    payMutation.mutate(appointment);
  }

  async function handleConfirmCancel(reason: string) {
    if (cancelingAppointments.length === 0) return;
    setIsCanceling(true);
    const results = await Promise.allSettled(
      cancelingAppointments.map((a) => cancelMutation.mutateAsync({ appointmentId: a.appointmentId, reason }))
    );
    setIsCanceling(false);

    const succeeded = results.filter((r) => r.status === "fulfilled").length;
    if (succeeded > 0) showSuccess("Booking canceled.");
    if (succeeded < results.length) showError({ message: "The booking could not be canceled." });

    setCancelingAppointments([]);
  }

  return (
    <>
      <AppointmentListHeader
        title="Bookings"
        state={list}
        help={
          <>
            <p>
              The Actions button on each row opens a menu: booking details, the expert&apos;s profile, joining the
              consultation room, viewing the medical record, and paying or cancelling when that applies.
            </p>
            <p>
              The consultation room opens shortly before the session starts; until then its link stays hidden. The
              Status column shows when a medical record is available.
            </p>
          </>
        }
      />
      <StatusFilterLinks state={list} counts={filterCounts} />
      <DateSortFilters state={list} />
      <AppointmentListSummary
        state={list}
        visibleCount={visibleAppointments.length}
        totalItems={bookingPage?.totalItems ?? 0}
      />

      <AppointmentTable
        state={list}
        counterpartLabel="Expert"
        visibleIds={visibleAppointments.map((a) => a.appointmentId)}
        totalPages={totalPages}
        isLoading={isLoading}
        isFetching={isFetching}
        emptyText={list.hasFilters ? "No bookings match these filters." : "You don't have any bookings yet."}
      >
        {visibleAppointments.map((appointment) => (
          <BookingTableRow
            key={appointment.appointmentId}
            appointment={appointment}
            columns={list.columns}
            onCancel={(a) => setCancelingAppointments([a])}
            onPayNow={handlePayNow}
            isPaying={
              (payMutation.isPending || isPaymentBrowserOpen) && payingAppointmentId === appointment.appointmentId
            }
          />
        ))}
      </AppointmentTable>

      <CancelAppointmentsDialog
        appointments={cancelingAppointments}
        counterpartName={expertName}
        counterpartFallback="the expert"
        onOpenChange={(open) => !open && setCancelingAppointments([])}
        onConfirm={handleConfirmCancel}
        isSubmitting={isCanceling}
      />
    </>
  );
}
