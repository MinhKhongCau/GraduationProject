"use client";

import { useState } from "react";
import Link from "next/link";
import { CreditCard, Eye, FileText, UserRound, XCircle } from "lucide-react";
import { useAppointmentMedicalRecord } from "@/hooks";
import { ROUTES } from "@/constants";
import { MedicalRecordDetailModal } from "@/components/medical-record/MedicalRecordDetailModal";
import { MedicalRecordLoadingModal } from "@/components/medical-record/MedicalRecordLoadingModal";
import {
  AppointmentDetailModal,
  AppointmentIdentityCell,
  AppointmentOptionalCells,
  AppointmentPriceInfo,
  AppointmentSlotInfo,
  AppointmentStatusInfo,
  MeetingRoomActionItem,
  RowActionItem,
  RowActionSeparator,
  RowActionsDropdown,
  isCancelable,
  shortBookingId,
  type OptionalColumn,
} from "@/components/appointment";
import type { AppointmentWithExpert } from "@/hooks";

interface BookingActionsProps {
  appointment: AppointmentWithExpert;
  onCancel: (appointment: AppointmentWithExpert) => void;
  onPayNow: (appointment: AppointmentWithExpert) => void;
  isPaying: boolean;
  triggerClassName?: string;
}

/** The "Thao tác" menu for one booking, plus the detail and medical-record modals it opens. */
function BookingActions({ appointment, onCancel, onPayNow, isPaying, triggerClassName }: BookingActionsProps) {
  const [recordOpen, setRecordOpen] = useState(false);
  const [appointmentDetailOpen, setAppointmentDetailOpen] = useState(false);
  // The record is only requested once the user opens it; the list itself only carries `hasMedicalRecord`.
  const { data: medicalRecord, isFetching: isLoadingRecord } = useAppointmentMedicalRecord(
    appointment.appointmentId,
    { enabled: recordOpen }
  );

  const canPay = appointment.statusLabel === "PENDING_PAYMENT";

  return (
    <>
      <RowActionsDropdown bookingId={appointment.appointmentId} triggerClassName={triggerClassName}>
        <RowActionItem onSelect={() => setAppointmentDetailOpen(true)}>
          <Eye className="h-4 w-4" aria-hidden="true" />
          Chi tiết
        </RowActionItem>
        <RowActionItem asChild>
          <Link href={ROUTES.PATIENT.EXPERT_DETAIL(appointment.expertId)}>
            <UserRound className="h-4 w-4" aria-hidden="true" />
            Hồ sơ chuyên gia
          </Link>
        </RowActionItem>
        <MeetingRoomActionItem appointment={appointment} />
        {appointment.hasMedicalRecord && (
          <RowActionItem onSelect={() => setRecordOpen(true)}>
            <FileText className="h-4 w-4" aria-hidden="true" />
            Xem bệnh án
          </RowActionItem>
        )}
        {canPay && (
          <RowActionItem disabled={isPaying} onSelect={() => onPayNow(appointment)}>
            <CreditCard className="h-4 w-4" aria-hidden="true" />
            {isPaying ? "Đang chuyển hướng..." : "Thanh toán ngay"}
          </RowActionItem>
        )}
        {isCancelable(appointment) && (
          <>
            <RowActionSeparator />
            <RowActionItem danger onSelect={() => onCancel(appointment)}>
              <XCircle className="h-4 w-4" aria-hidden="true" />
              Hủy lịch
            </RowActionItem>
          </>
        )}
      </RowActionsDropdown>

      <AppointmentDetailModal
        appointmentId={appointmentDetailOpen ? appointment.appointmentId : null}
        onClose={() => setAppointmentDetailOpen(false)}
      />
      {recordOpen && !medicalRecord ? (
        <MedicalRecordLoadingModal open={recordOpen} onOpenChange={setRecordOpen} isLoading={isLoadingRecord} />
      ) : (
        <MedicalRecordDetailModal
          open={recordOpen}
          onOpenChange={setRecordOpen}
          record={medicalRecord || null}
          isExpert={false}
        />
      )}
    </>
  );
}

export interface BookingTableRowProps {
  appointment: AppointmentWithExpert;
  columns: Record<OptionalColumn, boolean>;
  onCancel: (appointment: AppointmentWithExpert) => void;
  onPayNow: (appointment: AppointmentWithExpert) => void;
  isPaying: boolean;
}

export function BookingTableRow({ appointment, columns, onCancel, onPayNow, isPaying }: BookingTableRowProps) {
  const [appointmentDetailOpen, setAppointmentDetailOpen] = useState(false);

  return (
    <tr className="hover:bg-surface/60">
      <AppointmentIdentityCell appointment={appointment} onOpenDetail={() => setAppointmentDetailOpen(true)} />
      <td className="px-3 py-3 align-top text-sm font-medium text-foreground">
        {appointment.expertName ?? "Chuyên gia"}
      </td>
      <AppointmentOptionalCells appointment={appointment} columns={columns} />
      <td className="px-3 py-3 text-right align-top">
        <BookingActions appointment={appointment} onCancel={onCancel} onPayNow={onPayNow} isPaying={isPaying} />
      </td>
      <AppointmentDetailModal
        appointmentId={appointmentDetailOpen ? appointment.appointmentId : null}
        onClose={() => setAppointmentDetailOpen(false)}
      />
    </tr>
  );
}

export type BookingCardProps = Omit<BookingTableRowProps, "columns">;

/** Phone layout of one booking: expert, session time, status, amount and the action menu. */
export function BookingCard({ appointment, onCancel, onPayNow, isPaying }: BookingCardProps) {
  const [appointmentDetailOpen, setAppointmentDetailOpen] = useState(false);

  return (
    <article className="rounded-xl border border-border bg-background p-4 shadow-card">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="truncate text-base font-semibold text-foreground">{appointment.expertName ?? "Chuyên gia"}</h3>
          {appointment.specializationName && (
            <p className="truncate text-xs text-muted-foreground">{appointment.specializationName}</p>
          )}
        </div>
        <button
          type="button"
          onClick={() => setAppointmentDetailOpen(true)}
          className="shrink-0 rounded-md bg-primary-soft px-2 py-0.5 font-mono text-xs font-semibold text-primary-soft-text"
          aria-label={`Chi tiết lịch hẹn ${shortBookingId(appointment.appointmentId)}`}
        >
          {shortBookingId(appointment.appointmentId)}
        </button>
      </div>

      <div className="mt-3 rounded-lg bg-surface px-3 py-2.5 text-sm">
        <AppointmentSlotInfo appointment={appointment} />
      </div>

      <div className="mt-3">
        <AppointmentStatusInfo appointment={appointment} />
      </div>

      <div className="mt-3 flex items-end justify-between gap-3 border-t border-border pt-3">
        <div className="text-sm font-semibold tabular-nums text-foreground">
          <AppointmentPriceInfo appointment={appointment} />
        </div>
        <BookingActions
          appointment={appointment}
          onCancel={onCancel}
          onPayNow={onPayNow}
          isPaying={isPaying}
        />
      </div>

      <AppointmentDetailModal
        appointmentId={appointmentDetailOpen ? appointment.appointmentId : null}
        onClose={() => setAppointmentDetailOpen(false)}
      />
    </article>
  );
}
