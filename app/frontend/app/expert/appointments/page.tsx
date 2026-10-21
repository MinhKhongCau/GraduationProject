"use client";

import { useState } from "react";
import { AppointmentTableRow } from "./component/AppointmentTableRow";
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
import { useExpertAppointmentsPage, useExpertAppointmentCounts, useCancelBooking } from "@/hooks";
import { useErrorContext } from "@/context/ErrorContext";
import type { AppointmentWithPatient } from "@/hooks";

const patientName = (a: AppointmentWithPatient) => a.patientName;

export default function ExpertAppointmentsPage() {
  const list = useAppointmentListState();
  const [cancelingAppointments, setCancelingAppointments] = useState<AppointmentWithPatient[]>([]);
  const [isCanceling, setIsCanceling] = useState(false);
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

  async function handleConfirmCancel(reason: string) {
    if (cancelingAppointments.length === 0) return;
    setIsCanceling(true);
    const results = await Promise.allSettled(
      cancelingAppointments.map((a) => cancelMutation.mutateAsync({ appointmentId: a.appointmentId, reason }))
    );
    setIsCanceling(false);

    const succeeded = results.filter((r) => r.status === "fulfilled").length;
    if (succeeded > 0) showSuccess("Appointment canceled.");
    if (succeeded < results.length) showError({ message: "The appointment could not be canceled." });

    setCancelingAppointments([]);
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
            <p>
              The consultation room opens shortly before the session starts; until then its link stays hidden. The
              Status column shows which appointments already have a medical record.
            </p>
          </>
        }
      />
      <StatusFilterLinks state={list} counts={filterCounts} />
      <DateSortFilters state={list} />
      <AppointmentListSummary
        state={list}
        visibleCount={visibleAppointments.length}
        totalItems={appointmentPage?.totalItems ?? 0}
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
            onCancel={(a) => setCancelingAppointments([a])}
          />
        ))}
      </AppointmentTable>

      <CancelAppointmentsDialog
        appointments={cancelingAppointments}
        counterpartName={patientName}
        counterpartFallback="the patient"
        onOpenChange={(open) => !open && setCancelingAppointments([])}
        onConfirm={handleConfirmCancel}
        isSubmitting={isCanceling}
      />
    </>
  );
}
