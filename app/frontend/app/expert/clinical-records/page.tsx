"use client";

import { useState, useMemo } from "react";
import {
  FileText,
  Search,
  Calendar,
  AlertCircle,
  ShieldAlert,
  CheckCircle2,
  Edit3,
  Eye,
  User,
  HeartPulse,
} from "lucide-react";
import { Card, Button, Spinner } from "@/components/ui";
import { useMedicalRecords } from "@/hooks";
import { SaveMedicalRecordModal } from "@/components/medical-record/SaveMedicalRecordModal";
import { MedicalRecordDetailModal } from "@/components/medical-record/MedicalRecordDetailModal";
import type { MedicalRecord } from "@/types";

export default function ExpertClinicalRecordsPage() {
  const [search, setSearch] = useState("");
  const [selectedRecord, setSelectedRecord] = useState<MedicalRecord | null>(null);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);

  const { data, isLoading } = useMedicalRecords({ page: 0, size: 100 });
  const records = data?.items || [];

  const filteredRecords = useMemo(() => {
    if (!search.trim()) return records;
    const q = search.toLowerCase();
    return records.filter(
      (r) =>
        r.patientName?.toLowerCase().includes(q) ||
        r.diagnosis?.toLowerCase().includes(q) ||
        r.symptoms?.toLowerCase().includes(q) ||
        r.record_id?.toLowerCase().includes(q)
    );
  }, [records, search]);

  const upcomingFollowUps = useMemo(() => {
    const now = Date.now();
    return records.filter((r) => {
      const nextDate = r.next_appointment_date || r.nextAppointmentDate;
      return nextDate && nextDate >= now;
    }).length;
  }, [records]);

  const handleViewDetail = (record: MedicalRecord) => {
    setSelectedRecord(record);
    setDetailModalOpen(true);
  };

  const handleEditRecord = (record: MedicalRecord) => {
    setSelectedRecord(record);
    setEditModalOpen(true);
  };

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Hồ Sơ Bệnh Án & Bệnh Nhân</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Quản lý và theo dõi chẩn đoán, triệu chứng, hướng dẫn điều trị và lịch tái khám.
          </p>
        </div>
      </div>

      {/* Metrics Banner */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Card className="p-4 flex items-center gap-4">
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-primary-soft text-primary">
            <FileText className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Tổng số hồ sơ</p>
            <p className="text-2xl font-bold text-foreground">{records.length}</p>
          </div>
        </Card>

        <Card className="p-4 flex items-center gap-4">
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-success-soft text-success">
            <Calendar className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Lịch hẹn tái khám sắp tới</p>
            <p className="text-2xl font-bold text-foreground">{upcomingFollowUps}</p>
          </div>
        </Card>

        <Card className="p-4 flex items-center gap-4">
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-secondary/10 text-secondary">
            <HeartPulse className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Ca tư vấn đã hoàn thành</p>
            <p className="text-2xl font-bold text-foreground">{records.length}</p>
          </div>
        </Card>
      </div>

      {/* Search Filter */}
      <div className="relative max-w-md">
        <Search className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Tìm theo tên bệnh nhân, chẩn đoán, triệu chứng..."
          className="w-full rounded-2xl border border-border bg-surface py-2.5 pl-10 pr-4 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none"
        />
      </div>

      {/* List / Cards */}
      {isLoading ? (
        <div className="flex justify-center py-16">
          <Spinner className="h-7 w-7" />
        </div>
      ) : filteredRecords.length === 0 ? (
        <Card className="p-12 text-center">
          <FileText className="mx-auto h-12 w-12 text-muted-foreground/50 mb-3" />
          <h3 className="text-base font-bold text-foreground">Không tìm thấy hồ sơ bệnh án nào</h3>
          <p className="mt-1 text-sm text-muted-foreground max-w-md mx-auto">
            {search
              ? "Không có hồ sơ nào khớp với từ khóa tìm kiếm."
              : "Bạn chưa có hồ sơ bệnh án nào. Hồ sơ sẽ được tạo khi bạn thực hiện ghi chú cho các buổi khám của mình."}
          </p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {filteredRecords.map((record) => {
            const recordId = record.record_id || record.recordId || "";
            const apptId = record.appointment_id || record.appointmentId || "";
            const createdMs =
              record.created_at ||
              (typeof record.createdAt === "number" ? record.createdAt : Date.now());
            const nextDateMs = record.next_appointment_date || record.nextAppointmentDate;
            const actionsAvoid = record.actions_to_avoid || record.actionsToAvoid;
            const actionsTake = record.actions_to_take || record.actionsToTake;

            return (
              <Card
                key={recordId}
                className="p-5 flex flex-col gap-4 md:flex-row md:items-center md:justify-between hover:border-primary/40 transition-all duration-200"
              >
                <div className="space-y-2 flex-1">
                  <div className="flex flex-wrap items-center gap-2.5">
                    <div className="flex items-center gap-2 font-bold text-foreground">
                      <User className="h-4 w-4 text-primary" />
                      <span>{record.patientName || `Bệnh nhân (#${(record.patient_id || "").slice(0, 8)})`}</span>
                    </div>
                    <span className="rounded-md bg-surface px-2 py-0.5 text-xs text-muted-foreground">
                      {new Date(createdMs).toLocaleDateString("vi-VN")}
                    </span>
                    {nextDateMs && (
                      <span className="inline-flex items-center gap-1 rounded-md bg-primary-soft px-2 py-0.5 text-xs font-semibold text-primary">
                        <Calendar className="h-3 w-3" />
                        Tái khám: {new Date(nextDateMs).toLocaleDateString("vi-VN")}
                      </span>
                    )}
                  </div>

                  {/* Chẩn đoán */}
                  <div>
                    <span className="text-xs font-bold uppercase tracking-wider text-muted-foreground mr-2">
                      Chẩn đoán:
                    </span>
                    <span className="text-sm font-semibold text-foreground">
                      {record.diagnosis || "Chưa có chẩn đoán"}
                    </span>
                  </div>

                  {/* Triệu chứng tóm tắt */}
                  {record.symptoms && (
                    <p className="text-xs text-muted-foreground line-clamp-1">
                      <span className="font-semibold text-foreground/80">Triệu chứng:</span>{" "}
                      {record.symptoms}
                    </p>
                  )}

                  {/* Tags tóm tắt Cần tránh / Cần làm */}
                  <div className="flex flex-wrap gap-2 pt-1">
                    {actionsAvoid && (
                      <span className="inline-flex items-center gap-1 rounded-md bg-danger-soft/40 px-2 py-0.5 text-[11px] font-medium text-danger">
                        <ShieldAlert className="h-3 w-3" />
                        Có lưu ý CẦN TRÁNH
                      </span>
                    )}
                    {actionsTake && (
                      <span className="inline-flex items-center gap-1 rounded-md bg-success-soft/40 px-2 py-0.5 text-[11px] font-medium text-success">
                        <CheckCircle2 className="h-3 w-3" />
                        Có hướng dẫn CẦN LÀM
                      </span>
                    )}
                  </div>
                </div>

                {/* Nút hành động */}
                <div className="flex items-center gap-2 pt-2 md:pt-0 shrink-0">
                  <Button
                    size="sm"
                    variant="outline"
                    className="gap-1 text-xs"
                    onClick={() => handleViewDetail(record)}
                  >
                    <Eye className="h-3.5 w-3.5" />
                    Xem chi tiết
                  </Button>
                  <Button
                    size="sm"
                    className="gap-1 text-xs"
                    onClick={() => handleEditRecord(record)}
                  >
                    <Edit3 className="h-3.5 w-3.5" />
                    Chỉnh sửa
                  </Button>
                </div>
              </Card>
            );
          })}
        </div>
      )}

      {/* Modal chi tiết */}
      <MedicalRecordDetailModal
        open={detailModalOpen}
        onOpenChange={setDetailModalOpen}
        record={selectedRecord}
        onEdit={(rec) => {
          setSelectedRecord(rec);
          setEditModalOpen(true);
        }}
        isExpert
      />

      {/* Modal chỉnh sửa */}
      {selectedRecord && (
        <SaveMedicalRecordModal
          open={editModalOpen}
          onOpenChange={setEditModalOpen}
          appointmentId={selectedRecord.appointment_id || selectedRecord.appointmentId || ""}
          patientName={selectedRecord.patientName}
          initialData={selectedRecord}
          onSaved={(updated) => {
            setSelectedRecord(updated);
          }}
        />
      )}
    </div>
  );
}
