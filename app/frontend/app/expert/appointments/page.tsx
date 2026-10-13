"use client";

import { useState } from "react";
import { AppointmentTableRow } from "./component/AppointmentTableRow";
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
import { useExpertAppointmentsPage, useExpertAppointmentCounts, useCancelBooking } from "@/hooks";
import { useErrorContext } from "@/context/ErrorContext";
import type { AppointmentWithPatient } from "@/hooks";

const PATIENT: Counterpart<AppointmentWithPatient> = { label: "Patient", name: (a) => a.patientName };

export default function ExpertAppointmentsPage() {
  const list = useAppointmentListState();
  const [cancelingAppointments, setCancelingAppointments] = useState<AppointmentWithPatient[]>([]);
  const [isBulkCanceling, setIsBulkCanceling] = useState(false);
  const { showSuccess, showError } = useErrorContext();

  const { data: appointmentPage, isLoading, isFetching } = useExpertAppointmentsPage(list.params);
  const filterCounts = useExpertAppointmentCounts(STATUSES, list.dateRange);
  const cancelMutation = useCancelBooking();

  const visibleAppointments = appointmentPage?.items ?? [];
  const totalPages = appointmentPage?.totalPages ?? 0;

  // Cancelling the last rows of the last page (under a status filter) can leave `page` past the end.
  if (appointmentPage && !isFetching && list.page > 0 && list.page >= totalPages) {
    list.setPage(Math.max(totalPages - 1, 0));
  }

  const selectedAppointments = visibleAppointments.filter((a) => list.selectedIds.has(a.appointmentId));

  function handleApplyBulkAction() {
    if (selectedAppointments.length === 0) {
      showError({ message: "Select at least one appointment first." });
      return;
    }
    if (list.bulkAction === "download") {
      downloadBookingsCsv(selectedAppointments, PATIENT);
    } else if (list.bulkAction === "cancel") {
      const cancelable = selectedAppointments.filter(isCancelable);
      if (cancelable.length === 0) {
        showError({ message: "None of the selected appointments can be canceled." });
        return;
      }
      setCancelingAppointments(cancelable);
    }
  }

  function handleDownloadTickets() {
    const source = selectedAppointments.length > 0 ? selectedAppointments : visibleAppointments;
    if (downloadTicketsIcs(source, PATIENT) === 0)
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
    if (succeeded > 0) showSuccess(succeeded > 1 ? `${succeeded} appointments canceled.` : "Appointment canceled.");
    if (succeeded < results.length)
      showError({ message: `${results.length - succeeded} appointment(s) could not be canceled.` });

    setCancelingAppointments([]);
    list.clearSelection();
    list.setBulkAction("");
  }

  return (
    <>
      <AppointmentListHeader
        title="Appointments"
        state={list}
        help={
          <>
            <p>
              The Actions button on each row opens a menu: appointment details, joining the consultation room, viewing
              or writing the medical record, and cancelling when that applies.
            </p>
            <p>Select rows and use Bulk Actions to cancel or download several appointments at once.</p>
            <p>
              &quot;Download bookings&quot; and &quot;Download tickets&quot; use the appointments on the current page;
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
        totalItems={appointmentPage?.totalItems ?? 0}
        onApplyBulkAction={handleApplyBulkAction}
        onDownloadBookings={() => downloadBookingsCsv(visibleAppointments, PATIENT)}
        onDownloadTickets={handleDownloadTickets}
      />

      <AppointmentTable
        state={list}
        counterpartLabel="Patient"
        visibleIds={visibleAppointments.map((a) => a.appointmentId)}
        totalPages={totalPages}
        isLoading={isLoading}
        isFetching={isFetching}
        emptyText={list.hasFilters ? "No appointments match these filters." : "You don't have any appointments yet."}
      >
        {visibleAppointments.map((appointment) => (
          <AppointmentTableRow
            key={appointment.appointmentId}
            appointment={appointment}
            columns={list.columns}
            selected={list.selectedIds.has(appointment.appointmentId)}
            onSelectedChange={(selected) => list.toggleSelected(appointment.appointmentId, selected)}
            onCancel={(a) => setCancelingAppointments([a])}
          />
        ))}
      </AppointmentTable>

      <CancelAppointmentsDialog
        appointments={cancelingAppointments}
        counterpartName={PATIENT.name}
        counterpartFallback="the patient"
        onOpenChange={(open) => !open && setCancelingAppointments([])}
        onConfirm={handleConfirmCancel}
        isSubmitting={isBulkCanceling}
      />
    </>
  );
}
