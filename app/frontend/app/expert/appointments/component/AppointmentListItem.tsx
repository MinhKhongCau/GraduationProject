"use client";

import { useState } from "react";
import { CalendarClock, FileText, Video, ClipboardList } from "lucide-react";
import { Card, Button } from "@/components/ui";
import { useAppointmentMedicalRecord } from "@/hooks";
import { SaveMedicalRecordModal } from "@/components/medical-record/SaveMedicalRecordModal";
import { MedicalRecordDetailModal } from "@/components/medical-record/MedicalRecordDetailModal";
import type { Appointment, AppointmentStatus } from "@/types";

const STATUS_STYLES: Record<AppointmentStatus, string> = {
  PENDING_PAYMENT: "bg-warning-soft text-warning",
  CONFIRMED: "bg-success-soft text-success",
  CANCELLED: "bg-danger-soft text-danger",
  COMPLETED: "bg-primary-soft text-primary",
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
      <Card className="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between hover:border-primary/30 transition-all duration-200">
        <div className="flex items-start gap-3.5">
          <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary">
            <CalendarClock className="h-5 w-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <p className="text-sm font-bold text-foreground">
                {appointment.patientName || `Bệnh nhân (#${appointment.patientId.slice(0, 8)})`}
              </p>
              {medicalRecord && (
                <span className="inline-flex items-center gap-1 rounded-md bg-secondary/10 px-2 py-0.5 text-[11px] font-semibold text-secondary">
                  <FileText className="h-3 w-3" />
                  Đã có bệnh án
                </span>
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
          <span
            className={`rounded-lg px-2.5 py-1 text-[11px] font-bold uppercase tracking-wider ${STATUS_STYLES[appointment.statusLabel] || "bg-surface text-muted-foreground"}`}
          >
            {STATUS_TEXT[appointment.statusLabel] || appointment.statusLabel}
          </span>

          {canAddRecord && (
            <>
              {medicalRecord ? (
                <>
                  <Button
                    size="sm"
                    variant="outline"
                    className="gap-1.5 text-xs"
                    onClick={() => setDetailModalOpen(true)}
                  >
                    <FileText className="h-3.5 w-3.5 text-primary" />
                    Xem bệnh án
                  </Button>
                  <Button
                    size="sm"
                    className="gap-1.5 text-xs"
                    onClick={() => setSaveModalOpen(true)}
                  >
                    <ClipboardList className="h-3.5 w-3.5" />
                    Cập nhật bệnh án
                  </Button>
                </>
              ) : (
                <Button
                  size="sm"
                  className="gap-1.5 text-xs"
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
