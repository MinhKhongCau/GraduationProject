"use client";

import { useState } from "react";
import { CalendarClock, FileText, Video, ClipboardList } from "lucide-react";
import { Badge, Card, Button, type BadgeTone } from "@/components/ui";
import { useAppointmentMedicalRecord } from "@/hooks";
import { SaveMedicalRecordModal } from "@/components/medical-record/SaveMedicalRecordModal";
import { MedicalRecordDetailModal } from "@/components/medical-record/MedicalRecordDetailModal";
import type { Appointment, AppointmentStatus } from "@/types";

const STATUS_TONES: Record<AppointmentStatus, BadgeTone> = {
  PENDING_PAYMENT: "warning",
  CONFIRMED: "success",
  CANCELLED: "danger",
  COMPLETED: "primary",
};

const STATUS_TEXT: Record<AppointmentStatus, string> = {
  PENDING_PAYMENT: "Chờ thanh toán",
  CONFIRMED: "Đã xác nhận",
  CANCELLED: "Đã hủy",
  COMPLETED: "Đã hoàn thành",
};

export interface AppointmentWithPatient extends Appointment {
  patientName?: string;
  patientEmail?: string;
}

export function AppointmentListItem({ appointment }: { appointment: AppointmentWithPatient }) {
  const [saveModalOpen, setSaveModalOpen] = useState(false);
  const [detailModalOpen, setDetailModalOpen] = useState(false);

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
