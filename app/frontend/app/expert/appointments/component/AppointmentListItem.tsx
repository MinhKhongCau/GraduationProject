"use client";

import { useState } from "react";
import { CalendarClock, Eye, FileText, Video, ClipboardList } from "lucide-react";
import { Badge, Card, Button } from "@/components/ui";
import { useAppointmentMedicalRecord, type AppointmentWithPatient } from "@/hooks";
import { SaveMedicalRecordModal } from "@/components/medical-record/SaveMedicalRecordModal";
import { MedicalRecordDetailModal } from "@/components/medical-record/MedicalRecordDetailModal";
import {
  APPOINTMENT_STATUS_TEXT as STATUS_TEXT,
  APPOINTMENT_STATUS_TONES as STATUS_TONES,
  AppointmentDetailModal,
  formatAppointmentTime,
} from "@/components/appointment";

export function AppointmentListItem({ appointment }: { appointment: AppointmentWithPatient }) {
  const [saveModalOpen, setSaveModalOpen] = useState(false);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [appointmentDetailOpen, setAppointmentDetailOpen] = useState(false);

  const { data: medicalRecord } = useAppointmentMedicalRecord(appointment.appointmentId);

  const canAddRecord =
    appointment.statusLabel === "CONFIRMED" || appointment.statusLabel === "COMPLETED";

  return (
    <>
      <Card className="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between transition-colors duration-200 hover:border-primary/30">
        <div className="flex items-start gap-3.5">
          <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary-soft-text">
            <CalendarClock className="h-5 w-5" />
          </div>
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <p className="text-sm font-semibold text-foreground">
                {appointment.patientName || `Bệnh nhân (#${appointment.patientId.slice(0, 8)})`}
              </p>
              {medicalRecord && (
                <Badge tone="primary">
                  <FileText className="h-3 w-3" aria-hidden="true" />
                  Đã có bệnh án
                </Badge>
              )}
            </div>
            <p className="mt-0.5 text-xs font-medium text-foreground">
              {formatAppointmentTime(appointment.startTime, appointment.endTime)}
              {appointment.specializationName && (
                <span className="text-muted-foreground"> · {appointment.specializationName}</span>
              )}
            </p>
            <p className="text-xs text-muted-foreground mt-0.5">
              Thời gian đặt: {new Date(appointment.createdAt).toLocaleString("vi-VN")}
            </p>
            {appointment.meetingLink && appointment.statusLabel === "CONFIRMED" && (
              <a
                href={appointment.meetingLink}
                target="_blank"
                rel="noreferrer"
                className="mt-1.5 inline-flex items-center gap-1.5 text-xs font-semibold text-primary hover:underline"
              >
                <Video className="h-3.5 w-3.5" />
                Vào phòng tư vấn
              </a>
            )}
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2.5 sm:justify-end">
          <Badge tone={STATUS_TONES[appointment.statusLabel] ?? "neutral"}>
            {STATUS_TEXT[appointment.statusLabel] || appointment.statusLabel}
          </Badge>

          <Button size="sm" variant="ghost" onClick={() => setAppointmentDetailOpen(true)}>
            <Eye className="h-3.5 w-3.5" />
            Chi tiết
          </Button>

          {canAddRecord && (
            <>
              {medicalRecord ? (
                <>
                  <Button
                    size="sm"
                    variant="outline"
                                        onClick={() => setDetailModalOpen(true)}
                  >
                    <FileText className="h-3.5 w-3.5 text-primary" />
                    Xem bệnh án
                  </Button>
                  <Button
                    size="sm"
                                        onClick={() => setSaveModalOpen(true)}
                  >
                    <ClipboardList className="h-3.5 w-3.5" />
                    Cập nhật bệnh án
                  </Button>
                </>
              ) : (
                <Button
                  size="sm"
                                    onClick={() => setSaveModalOpen(true)}
                >
                  <ClipboardList className="h-3.5 w-3.5" />
                  Ghi chú bệnh án
                </Button>
              )}
            </>
          )}
        </div>
      </Card>

      <AppointmentDetailModal
        appointmentId={appointmentDetailOpen ? appointment.appointmentId : null}
        onClose={() => setAppointmentDetailOpen(false)}
      />

      {/* Modal chỉnh sửa / ghi chú bệnh án */}
      <SaveMedicalRecordModal
        open={saveModalOpen}
        onOpenChange={setSaveModalOpen}
        appointmentId={appointment.appointmentId}
        patientName={appointment.patientName}
        initialData={medicalRecord}
      />

      {/* Modal xem chi tiết bệnh án */}
      <MedicalRecordDetailModal
        open={detailModalOpen}
        onOpenChange={setDetailModalOpen}
        record={medicalRecord || null}
        onEdit={() => setSaveModalOpen(true)}
        isExpert
      />
    </>
  );
}
