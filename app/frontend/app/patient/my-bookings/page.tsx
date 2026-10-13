"use client";

import { useState } from "react";
import { BookingTableRow } from "./component/BookingTableRow";
import {
  AppointmentListHeader,
  AppointmentListToolbar,
  AppointmentTable,
  CancelAppointmentsDialog,
  DateSortFilters,
  STATUSES,
  StatusFilterLinks,
  downloadBookingsCsv,
  downloadTicketsIcs,
  isCancelable,
  useAppointmentListState,
  type Counterpart,
} from "@/components/appointment";
import { useMyBookingsPage, useMyBookingCounts, useCancelBooking, usePaymentRedirect } from "@/hooks";
import { useApiMutation } from "@/hooks";
import { paymentApi } from "@/api";
import { CHANNELING_FEE } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type { AppointmentWithExpert } from "@/hooks";

const EXPERT: Counterpart<AppointmentWithExpert> = { label: "Expert", name: (a) => a.expertName };

export default function MyBookingsPage() {
  const list = useAppointmentListState();
  const [cancelingAppointments, setCancelingAppointments] = useState<AppointmentWithExpert[]>([]);
  const [payingAppointmentId, setPayingAppointmentId] = useState<string | null>(null);
  const [isBulkCanceling, setIsBulkCanceling] = useState(false);
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

  const selectedAppointments = visibleAppointments.filter((a) => list.selectedIds.has(a.appointmentId));

  function handlePayNow(appointment: AppointmentWithExpert) {
    setPayingAppointmentId(appointment.appointmentId);
    payMutation.mutate(appointment);
  }

  function handleApplyBulkAction() {
    if (selectedAppointments.length === 0) {
      showError({ message: "Select at least one booking first." });
      return;
    }
    if (list.bulkAction === "download") {
      downloadBookingsCsv(selectedAppointments, EXPERT);
    } else if (list.bulkAction === "cancel") {
      const cancelable = selectedAppointments.filter(isCancelable);
      if (cancelable.length === 0) {
        showError({ message: "None of the selected bookings can be canceled." });
        return;
      }
      setCancelingAppointments(cancelable);
    }
  }

  function handleDownloadTickets() {
    const source = selectedAppointments.length > 0 ? selectedAppointments : visibleAppointments;
    if (downloadTicketsIcs(source, EXPERT) === 0)
      showError({ message: "No confirmed sessions to download tickets for." });
  }

  async function handleConfirmCancel(reason: string) {
    if (cancelingAppointments.length === 0) return;
    setIsBulkCanceling(true);
    const results = await Promise.allSettled(
      cancelingAppointments.map((a) => cancelMutation.mutateAsync({ appointmentId: a.appointmentId, reason }))
    );
    setIsBulkCanceling(false);

    const succeeded = results.filter((r) => r.status === "fulfilled").length;
    if (succeeded > 0) showSuccess(succeeded > 1 ? `${succeeded} bookings canceled.` : "Booking canceled.");
    if (succeeded < results.length)
      showError({ message: `${results.length - succeeded} booking(s) could not be canceled.` });

    setCancelingAppointments([]);
    list.clearSelection();
    list.setBulkAction("");
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
              consultation room, and paying or cancelling when that applies.
            </p>
            <p>Select rows and use Bulk Actions to cancel or download several bookings at once.</p>
            <p>
              &quot;Download bookings&quot; and &quot;Download tickets&quot; use the bookings on the current page;
              tickets are a calendar file (.ics) of confirmed sessions.
            </p>
          </>
        }
      />
      <StatusFilterLinks state={list} counts={filterCounts} />
      <DateSortFilters state={list} />
      <AppointmentListToolbar
        state={list}
        visibleCount={visibleAppointments.length}
        selectedCount={selectedAppointments.length}
        totalItems={bookingPage?.totalItems ?? 0}
        onApplyBulkAction={handleApplyBulkAction}
        onDownloadBookings={() => downloadBookingsCsv(visibleAppointments, EXPERT)}
        onDownloadTickets={handleDownloadTickets}
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
            selected={list.selectedIds.has(appointment.appointmentId)}
            onSelectedChange={(selected) => list.toggleSelected(appointment.appointmentId, selected)}
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
        counterpartName={EXPERT.name}
        counterpartFallback="the expert"
        onOpenChange={(open) => !open && setCancelingAppointments([])}
        onConfirm={handleConfirmCancel}
        isSubmitting={isBulkCanceling}
      />
    </>
  );
}
