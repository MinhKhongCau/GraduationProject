"use client";

import { useState } from "react";
import { ClipboardList, Eye, FileText, Video, XCircle } from "lucide-react";
import { useAppointmentMedicalRecord, type AppointmentWithPatient } from "@/hooks";
import { SaveMedicalRecordModal } from "@/components/medical-record/SaveMedicalRecordModal";
import { MedicalRecordDetailModal } from "@/components/medical-record/MedicalRecordDetailModal";
import {
  AppointmentDetailModal,
  AppointmentIdentityCells,
  AppointmentOptionalCells,
  RowActionItem,
  RowActionSeparator,
  RowActionsMenu,
  isCancelable,
  type OptionalColumn,
} from "@/components/appointment";

export interface AppointmentTableRowProps {
  appointment: AppointmentWithPatient;
  columns: Record<OptionalColumn, boolean>;
  selected: boolean;
  onSelectedChange: (selected: boolean) => void;
  onCancel: (appointment: AppointmentWithPatient) => void;
}

export function AppointmentTableRow({
  appointment,
  columns,
  selected,
  onSelectedChange,
  onCancel,
}: AppointmentTableRowProps) {
  const [saveModalOpen, setSaveModalOpen] = useState(false);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [appointmentDetailOpen, setAppointmentDetailOpen] = useState(false);
  const { data: medicalRecord } = useAppointmentMedicalRecord(appointment.appointmentId);

  const isConfirmed = appointment.statusLabel === "CONFIRMED";
  const canWriteRecord = isConfirmed || appointment.statusLabel === "COMPLETED";

  return (
    <tr className={selected ? "bg-primary-soft/40" : "hover:bg-surface/60"}>
      <AppointmentIdentityCells
        appointment={appointment}
        selected={selected}
        onSelectedChange={onSelectedChange}
        onOpenDetail={() => setAppointmentDetailOpen(true)}
        hasMedicalRecord={!!medicalRecord}
      />
      <td className="px-3 py-3 align-top text-sm">
        <p className="font-medium text-foreground">
          {appointment.patientName || `Bệnh nhân (#${appointment.patientId.slice(0, 8)})`}
        </p>
        {appointment.patientEmail && <p className="text-xs text-muted-foreground">{appointment.patientEmail}</p>}
      </td>
      <AppointmentOptionalCells appointment={appointment} columns={columns} />

      <RowActionsMenu bookingId={appointment.appointmentId}>
        <RowActionItem onSelect={() => setAppointmentDetailOpen(true)}>
          <Eye className="h-4 w-4" aria-hidden="true" />
          Chi tiết
        </RowActionItem>
        {isConfirmed &&
          (appointment.meetingLink ? (
            <RowActionItem asChild>
              <a href={appointment.meetingLink} target="_blank" rel="noreferrer">
                <Video className="h-4 w-4" aria-hidden="true" />
                Vào phòng tư vấn
              </a>
            </RowActionItem>
          ) : (
            <RowActionItem disabled>
              <Video className="h-4 w-4" aria-hidden="true" />
              Phòng tư vấn chưa sẵn sàng
            </RowActionItem>
          ))}
        {medicalRecord && (
          <RowActionItem onSelect={() => setDetailModalOpen(true)}>
            <FileText className="h-4 w-4" aria-hidden="true" />
            Xem bệnh án
          </RowActionItem>
        )}
        {canWriteRecord && (
          <RowActionItem onSelect={() => setSaveModalOpen(true)}>
            <ClipboardList className="h-4 w-4" aria-hidden="true" />
            {medicalRecord ? "Cập nhật bệnh án" : "Ghi chú bệnh án"}
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
      <SaveMedicalRecordModal
        open={saveModalOpen}
        onOpenChange={setSaveModalOpen}
        appointmentId={appointment.appointmentId}
        patientName={appointment.patientName}
        initialData={medicalRecord}
      />
      <MedicalRecordDetailModal
        open={detailModalOpen}
        onOpenChange={setDetailModalOpen}
        record={medicalRecord || null}
        onEdit={() => setSaveModalOpen(true)}
        isExpert
      />
    </tr>
  );
}
