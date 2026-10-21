"use client";

import { useState } from "react";
import { ClipboardList, Eye, FileText, XCircle } from "lucide-react";
import { useAppointmentMedicalRecord, type AppointmentWithPatient } from "@/hooks";
import { SaveMedicalRecordModal } from "@/components/medical-record/SaveMedicalRecordModal";
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

export interface AppointmentTableRowProps {
  appointment: AppointmentWithPatient;
  columns: Record<OptionalColumn, boolean>;
  onCancel: (appointment: AppointmentWithPatient) => void;
}

export function AppointmentTableRow({ appointment, columns, onCancel }: AppointmentTableRowProps) {
  const [saveModalOpen, setSaveModalOpen] = useState(false);
  const [recordOpen, setRecordOpen] = useState(false);
  const [appointmentDetailOpen, setAppointmentDetailOpen] = useState(false);

  const hasRecord = !!appointment.hasMedicalRecord;
  // Load the record only when it is viewed, or edited when one already exists (to prefill the form).
  const needsRecord = recordOpen || (saveModalOpen && hasRecord);
  const { data: medicalRecord, isFetching: isLoadingRecord } = useAppointmentMedicalRecord(
    appointment.appointmentId,
    { enabled: needsRecord }
  );
  const recordPending = needsRecord && !medicalRecord;

  const canWriteRecord = appointment.statusLabel === "CONFIRMED" || appointment.statusLabel === "COMPLETED";

  return (
    <tr className="hover:bg-surface/60">
      <AppointmentIdentityCell appointment={appointment} onOpenDetail={() => setAppointmentDetailOpen(true)} />
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
        <MeetingRoomActionItem appointment={appointment} />
        {hasRecord && (
          <RowActionItem onSelect={() => setRecordOpen(true)}>
            <FileText className="h-4 w-4" aria-hidden="true" />
            Xem bệnh án
          </RowActionItem>
        )}
        {canWriteRecord && (
          <RowActionItem onSelect={() => setSaveModalOpen(true)}>
            <ClipboardList className="h-4 w-4" aria-hidden="true" />
            {hasRecord ? "Cập nhật bệnh án" : "Ghi chú bệnh án"}
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
      {recordPending ? (
        <MedicalRecordLoadingModal
          open
          onOpenChange={(open) => {
            if (!open) {
              setRecordOpen(false);
              setSaveModalOpen(false);
            }
          }}
          isLoading={isLoadingRecord}
        />
      ) : (
        <>
          <SaveMedicalRecordModal
            open={saveModalOpen}
            onOpenChange={setSaveModalOpen}
            appointmentId={appointment.appointmentId}
            patientName={appointment.patientName}
            initialData={medicalRecord}
          />
          <MedicalRecordDetailModal
            open={recordOpen}
            onOpenChange={setRecordOpen}
            record={medicalRecord || null}
            onEdit={() => setSaveModalOpen(true)}
            isExpert
          />
        </>
      )}
    </tr>
  );
}
