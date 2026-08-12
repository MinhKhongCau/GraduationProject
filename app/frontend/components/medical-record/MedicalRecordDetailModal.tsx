"use client";

import { Modal, Button } from "@/components/ui";
import type { MedicalRecord } from "@/types";
import {
  AlertCircle,
  Calendar,
  CheckCircle2,
  FileText,
  HeartPulse,
  ShieldAlert,
  Sparkles,
  User,
  Clock,
  ExternalLink,
} from "lucide-react";

interface MedicalRecordDetailModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  record: MedicalRecord | null;
  onEdit?: (record: MedicalRecord) => void;
  isExpert?: boolean;
}

export function MedicalRecordDetailModal({
  open,
  onOpenChange,
  record,
  onEdit,
  isExpert,
}: MedicalRecordDetailModalProps) {
  if (!record) return null;

  const createdAtMs = record.created_at || (typeof record.createdAt === "number" ? record.createdAt : Date.now());
  const nextDateMs = record.next_appointment_date || record.nextAppointmentDate;
  const actionsAvoid = record.actions_to_avoid || record.actionsToAvoid;
  const actionsTake = record.actions_to_take || record.actionsToTake;
  const treatment = record.treatment_plan || record.treatmentPlan;
  const nextNote = record.next_appointment_note || record.nextAppointmentNote;
  const notes = record.expert_notes || record.expertNotes;

  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Chi tiết Hồ sơ Bệnh án"
      description={`Mã hồ sơ: ${record.record_id || record.recordId || "N/A"}`}
      size="2xl"
    >
      <div className="space-y-5 pt-2">
        {/* Header thông tin người tham gia & ngày khám */}
        <div className="grid grid-cols-1 gap-3 rounded-2xl bg-surface p-4 sm:grid-cols-2">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary">
              <User className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">
                {isExpert ? "Bệnh nhân" : "Chuyên gia / Bác sĩ"}
              </p>
              <p className="text-sm font-bold text-foreground">
                {isExpert
                  ? record.patientName || "Bệnh nhân"
                  : record.expertName || "Chuyên gia tâm lý"}
              </p>
              {record.patientEmail && isExpert && (
                <p className="text-xs text-muted-foreground">{record.patientEmail}</p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-3 sm:justify-end">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-surface-raised text-muted-foreground">
              <Clock className="h-5 w-5" />
            </div>
            <div className="text-left sm:text-right">
              <p className="text-xs text-muted-foreground">Thời gian khám</p>
              <p className="text-sm font-semibold text-foreground">
                {new Date(createdAtMs).toLocaleDateString("vi-VN", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </p>
            </div>
          </div>
        </div>

        {/* Tình trạng bệnh / Chẩn đoán */}
        <div className="rounded-2xl border border-primary/20 bg-primary-soft/30 p-4">
          <div className="flex items-center gap-2 mb-1.5">
            <HeartPulse className="h-5 w-5 text-primary" />
            <span className="text-xs font-bold uppercase tracking-wider text-primary">
              Chẩn đoán & Tình trạng bệnh
            </span>
          </div>
          <p className="text-base font-bold text-foreground">{record.diagnosis || "Chưa ghi nhận"}</p>
        </div>

        {/* Triệu chứng lâm sàng */}
        {record.symptoms && (
          <div>
            <div className="flex items-center gap-1.5 mb-1.5 text-xs font-bold text-foreground">
              <AlertCircle className="h-4 w-4 text-warning" />
              Triệu chứng lâm sàng
            </div>
            <div className="rounded-xl border border-border bg-surface p-3.5 text-sm text-foreground whitespace-pre-line leading-relaxed">
              {record.symptoms}
            </div>
          </div>
        )}

        {/* Cần tránh & Cần làm */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {actionsAvoid && (
            <div className="rounded-xl border border-danger/20 bg-danger-soft/20 p-4 space-y-2">
              <div className="flex items-center gap-1.5 text-xs font-bold text-danger">
                <ShieldAlert className="h-4 w-4" />
                Các hành động CẦN TRÁNH
              </div>
              <p className="text-sm text-foreground whitespace-pre-line leading-relaxed">
                {actionsAvoid}
              </p>
            </div>
          )}

          {actionsTake && (
            <div className="rounded-xl border border-success/20 bg-success-soft/20 p-4 space-y-2">
              <div className="flex items-center gap-1.5 text-xs font-bold text-success">
                <CheckCircle2 className="h-4 w-4" />
                Các hành động CẦN LÀM / Lời dặn
              </div>
              <p className="text-sm text-foreground whitespace-pre-line leading-relaxed">
                {actionsTake}
              </p>
            </div>
          )}
        </div>

        {/* Phác đồ điều trị */}
        {treatment && (
          <div>
            <div className="flex items-center gap-1.5 mb-1.5 text-xs font-bold text-foreground">
              <FileText className="h-4 w-4 text-secondary" />
              Phác đồ / Hướng can thiệp
            </div>
            <div className="rounded-xl border border-border bg-surface p-3.5 text-sm text-foreground whitespace-pre-line leading-relaxed">
              {treatment}
            </div>
          </div>
        )}

        {/* Lịch hẹn tiếp theo */}
        {(nextDateMs || nextNote) && (
          <div className="rounded-xl border border-primary/20 bg-surface p-4 flex items-start gap-3.5">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary">
              <Calendar className="h-5 w-5" />
            </div>
            <div className="space-y-0.5">
              <span className="text-xs font-bold uppercase tracking-wider text-primary">
                Buổi gặp tiếp theo (Tái khám)
              </span>
              {nextDateMs && (
                <p className="text-sm font-bold text-foreground">
                  {new Date(nextDateMs).toLocaleDateString("vi-VN", {
                    weekday: "long",
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  })}
                </p>
              )}
              {nextNote && <p className="text-xs text-muted-foreground">{nextNote}</p>}
            </div>
          </div>
        )}

        {/* Ghi chú thêm của chuyên gia */}
        {notes && (
          <div>
            <div className="flex items-center gap-1.5 mb-1.5 text-xs font-bold text-muted-foreground">
              <Sparkles className="h-4 w-4" />
              Ghi chú thêm từ Chuyên gia
            </div>
            <div className="rounded-xl border border-border bg-surface p-3.5 text-xs text-muted-foreground italic whitespace-pre-line">
              {notes}
            </div>
          </div>
        )}

        {/* Nút hành động */}
        <div className="flex items-center justify-end gap-3 pt-2">
          {isExpert && onEdit && (
            <Button
              variant="outline"
              onClick={() => {
                onOpenChange(false);
                onEdit(record);
              }}
            >
              Chỉnh sửa hồ sơ
            </Button>
          )}
          <Button onClick={() => onOpenChange(false)}>Đóng</Button>
        </div>
      </div>
    </Modal>
  );
}
