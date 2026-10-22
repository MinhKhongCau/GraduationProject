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
  MeetingRoomActionItem,
  RowActionItem,
  RowActionSeparator,
  RowActionsMenu,
  isCancelable,
  type OptionalColumn,
} from "@/components/appointment";
import type { AppointmentWithExpert } from "@/hooks";

export interface BookingTableRowProps {
  appointment: AppointmentWithExpert;
  columns: Record<OptionalColumn, boolean>;
  onCancel: (appointment: AppointmentWithExpert) => void;
  onPayNow: (appointment: AppointmentWithExpert) => void;
  isPaying: boolean;
}

export function BookingTableRow({ appointment, columns, onCancel, onPayNow, isPaying }: BookingTableRowProps) {
  const [recordOpen, setRecordOpen] = useState(false);
  const [appointmentDetailOpen, setAppointmentDetailOpen] = useState(false);
  // The record is only requested once the user opens it; the list itself only carries `hasMedicalRecord`.
  const { data: medicalRecord, isFetching: isLoadingRecord } = useAppointmentMedicalRecord(
    appointment.appointmentId,
    { enabled: recordOpen }
  );

  const canPay = appointment.statusLabel === "PENDING_PAYMENT";

  return (
    <tr className="hover:bg-surface/60">
      <AppointmentIdentityCell appointment={appointment} onOpenDetail={() => setAppointmentDetailOpen(true)} />
      <td className="px-3 py-3 align-top text-sm font-medium text-foreground">
        {appointment.expertName ?? "Chuyên gia"}
      </td>
      <AppointmentOptionalCells appointment={appointment} columns={columns} />

      <RowActionsMenu bookingId={appointment.appointmentId}>
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
      </RowActionsMenu>

      <AppointmentDetailModal
        appointmentId={appointmentDetailOpen ? appointment.appointmentId : null}
        onClose={() => setAppointmentDetailOpen(false)}
      />
      {recordOpen && !medicalRecord ? (
        <MedicalRecordLoadingModal
          open={recordOpen}
          onOpenChange={setRecordOpen}
          isLoading={isLoadingRecord}
        />
      ) : (
        <MedicalRecordDetailModal
          open={recordOpen}
          onOpenChange={setRecordOpen}
          record={medicalRecord || null}
          isExpert={false}
        />
      )}
    </tr>
  );
}
