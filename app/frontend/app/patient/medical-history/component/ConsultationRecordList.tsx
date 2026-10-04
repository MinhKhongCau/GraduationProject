"use client";

import { useState } from "react";
import {
  FileText,
  Calendar,
  AlertCircle,
  ShieldAlert,
  CheckCircle2,
  Eye,
  User,
  HeartPulse,
  Clock,
} from "lucide-react";
import { Card, Button, Badge, Spinner } from "@/components/ui";
import { useMedicalRecords } from "@/hooks";
import { MedicalRecordDetailModal } from "@/components/medical-record/MedicalRecordDetailModal";
import type { MedicalRecord } from "@/types";

export function ConsultationRecordList() {
  const [selectedRecord, setSelectedRecord] = useState<MedicalRecord | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);

  const { data, isLoading } = useMedicalRecords({ page: 0, size: 100 });
  const records = data?.items || [];

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner className="h-6 w-6" />
      </div>
    );
  }

  if (records.length === 0) {
    return (
      <Card className="p-8 text-center">
        <HeartPulse className="mx-auto h-10 w-10 text-muted-foreground/40 mb-3" />
        <h3 className="text-sm font-semibold text-foreground">Chưa có hồ sơ bệnh án từ chuyên gia</h3>
        <p className="mt-1 text-xs text-muted-foreground max-w-sm mx-auto">
          Sau mỗi buổi tư vấn hoặc khám bệnh với Chuyên gia / Bác sĩ, kết luận chẩn đoán và hướng dẫn điều trị sẽ xuất hiện tại đây.
        </p>
      </Card>
    );
  }

  return (
    <>
      <div className="space-y-3">
        {records.map((record) => {
          const recordId = record.record_id || record.recordId || "";
          const createdMs =
            record.created_at ||
            (typeof record.createdAt === "number" ? record.createdAt : Date.now());
          const nextDateMs = record.next_appointment_date || record.nextAppointmentDate;
          const actionsAvoid = record.actions_to_avoid || record.actionsToAvoid;
          const actionsTake = record.actions_to_take || record.actionsToTake;

          return (
            <Card
              key={recordId}
              className="flex flex-col gap-3 p-5 transition-colors duration-200 hover:border-primary/30 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="space-y-1.5 flex-1">
                <div className="flex flex-wrap items-center gap-2">
                  <div className="flex items-center gap-1.5 text-xs font-bold text-primary">
                    <User className="h-3.5 w-3.5" />
                    <span>{record.expertName || "Chuyên gia tâm lý"}</span>
                  </div>
                  <span className="text-xs text-muted-foreground">•</span>
                  <span className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Clock className="h-3 w-3" />
                    {new Date(createdMs).toLocaleDateString("vi-VN", {
                      day: "2-digit",
                      month: "2-digit",
                      year: "numeric",
                    })}
                  </span>
                  {nextDateMs && (
                    <Badge tone="primary">
                      <Calendar className="h-3 w-3" aria-hidden="true" />
                      Tái khám: {new Date(nextDateMs).toLocaleDateString("vi-VN")}
                    </Badge>
                  )}
                </div>

                {/* Chẩn đoán */}
                <p className="text-sm font-semibold text-foreground">
                  {record.diagnosis || "Chưa ghi nhận chẩn đoán"}
                </p>

                {/* Triệu chứng tóm tắt */}
                {record.symptoms && (
                  <p className="text-xs text-muted-foreground line-clamp-1">
                    <span className="font-semibold">Triệu chứng:</span> {record.symptoms}
                  </p>
                )}

                {/* Gợi ý Cần tránh / Cần làm */}
                <div className="flex flex-wrap gap-2 pt-1">
                  {actionsAvoid && (
                    <Badge tone="danger">
                      <ShieldAlert className="h-3 w-3" aria-hidden="true" />
                      Có dặn dò CẦN TRÁNH
                    </Badge>
                  )}
                  {actionsTake && (
                    <Badge tone="success">
                      <CheckCircle2 className="h-3 w-3" aria-hidden="true" />
                      Có hướng dẫn CẦN LÀM
                    </Badge>
                  )}
                </div>
              </div>

              <div className="pt-2 sm:pt-0 shrink-0">
                <Button
                  size="sm"
                  variant="outline"
                  className="w-full sm:w-auto"
                  onClick={() => {
                    setSelectedRecord(record);
                    setDetailOpen(true);
                  }}
                >
                  <Eye className="h-3.5 w-3.5 text-primary" />
                  Xem bệnh án chi tiết
                </Button>
              </div>
            </Card>
          );
        })}
      </div>

      {/* Modal chi tiết */}
      <MedicalRecordDetailModal
        open={detailOpen}
        onOpenChange={setDetailOpen}
        record={selectedRecord}
        isExpert={false}
      />
    </>
  );
}
