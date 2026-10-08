"use client";

import { useState } from "react";
import { CalendarClock, Eye, FileText, Video } from "lucide-react";
import { Card, Button, Badge } from "@/components/ui";
import { useAppointmentMedicalRecord } from "@/hooks";
import { MedicalRecordDetailModal } from "@/components/medical-record/MedicalRecordDetailModal";
import {
  APPOINTMENT_STATUS_TEXT as STATUS_TEXT,
  APPOINTMENT_STATUS_TONES as STATUS_TONES,
  AppointmentDetailModal,
  formatAppointmentTime,
} from "@/components/appointment";
import type { AppointmentWithExpert } from "@/hooks";

export interface BookingListItemProps {
  appointment: AppointmentWithExpert;
  onCancel: (appointment: AppointmentWithExpert) => void;
  onPayNow: (appointment: AppointmentWithExpert) => void;
  isPaying: boolean;
}

export function BookingListItem({ appointment, onCancel, onPayNow, isPaying }: BookingListItemProps) {
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [appointmentDetailOpen, setAppointmentDetailOpen] = useState(false);
  const { data: medicalRecord } = useAppointmentMedicalRecord(appointment.appointmentId);

  const canCancel =
    appointment.statusLabel === "CONFIRMED" || appointment.statusLabel === "PENDING_PAYMENT";
  const canPay = appointment.statusLabel === "PENDING_PAYMENT";

  return (
    <>
      <Card className="flex flex-col gap-3 p-5 sm:flex-row sm:items-center sm:justify-between transition-colors duration-200 hover:border-primary/30">
        <div className="flex items-start gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary-soft-text">
            <CalendarClock className="h-5 w-5" />
          </div>
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <p className="text-sm font-semibold text-foreground">{appointment.expertName ?? "Chuyên gia"}</p>
              {medicalRecord && (
                <Badge tone="primary">
                  <FileText className="h-3 w-3" aria-hidden="true" />
                  Đã có bệnh án
                </Badge>
              )}
            </div>
            <p className="mt-1 text-xs font-medium text-foreground">
              {formatAppointmentTime(appointment.startTime, appointment.endTime)}
              {appointment.specializationName && (
                <span className="text-muted-foreground"> · {appointment.specializationName}</span>
              )}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              Đặt lịch lúc {new Date(appointment.createdAt).toLocaleString("vi-VN")}
              {appointment.price ? ` · ${appointment.price.toLocaleString("vi-VN")} đ` : ""}
            </p>
            {appointment.statusLabel === "CANCELLED" && appointment.cancellationReason && (
              <p className="mt-1 text-xs text-muted-foreground">Lý do hủy: {appointment.cancellationReason}</p>
            )}
            {appointment.statusLabel === "CONFIRMED" && appointment.meetingLink && (
              <a
                href={appointment.meetingLink}
                target="_blank"
                rel="noreferrer"
                className="mt-1 inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline"
              >
                <Video className="h-3 w-3" />
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

          {medicalRecord && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => setDetailModalOpen(true)}
            >
              <FileText className="h-3.5 w-3.5 text-primary" />
              Xem bệnh án
            </Button>
          )}

          {canPay && (
            <Button size="sm" onClick={() => onPayNow(appointment)} disabled={isPaying}>
              {isPaying ? "Đang chuyển hướng..." : "Thanh toán ngay"}
            </Button>
          )}

          {canCancel && (
            <Button size="sm" variant="outline" onClick={() => onCancel(appointment)}>
              Hủy lịch
            </Button>
          )}
        </div>
      </Card>

      <AppointmentDetailModal
        appointmentId={appointmentDetailOpen ? appointment.appointmentId : null}
        onClose={() => setAppointmentDetailOpen(false)}
      />

      {/* Modal xem chi tiết bệnh án */}
      <MedicalRecordDetailModal
        open={detailModalOpen}
        onOpenChange={setDetailModalOpen}
        record={medicalRecord || null}
        isExpert={false}
      />
    </>
  );
}
