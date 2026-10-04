"use client";

import { useEffect, useState } from "react";
import { Modal, Button, Input, Textarea } from "@/components/ui";
import { useSaveMedicalRecord } from "@/hooks";
import { useErrorContext } from "@/context/ErrorContext";
import type { MedicalRecord, SaveMedicalRecordRequest } from "@/types";
import { AlertCircle, Calendar, CheckCircle2, FileText, HeartPulse, ShieldAlert, Sparkles } from "lucide-react";

interface SaveMedicalRecordModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  appointmentId: string;
  patientName?: string;
  initialData?: MedicalRecord | null;
  onSaved?: (record: MedicalRecord) => void;
}

export function SaveMedicalRecordModal({
  open,
  onOpenChange,
  appointmentId,
  patientName,
  initialData,
  onSaved,
}: SaveMedicalRecordModalProps) {
  const { showSuccess, showError } = useErrorContext();
  const saveMutation = useSaveMedicalRecord();

  const [diagnosis, setDiagnosis] = useState("");
  const [symptoms, setSymptoms] = useState("");
  const [actionsToAvoid, setActionsToAvoid] = useState("");
  const [actionsToTake, setActionsToTake] = useState("");
  const [treatmentPlan, setTreatmentPlan] = useState("");
  const [nextAppointmentDate, setNextAppointmentDate] = useState("");
  const [nextAppointmentNote, setNextAppointmentNote] = useState("");
  const [expertNotes, setExpertNotes] = useState("");

  useEffect(() => {
    if (initialData) {
      setDiagnosis(initialData.diagnosis || "");
      setSymptoms(initialData.symptoms || "");
      setActionsToAvoid(initialData.actions_to_avoid || initialData.actionsToAvoid || "");
      setActionsToTake(initialData.actions_to_take || initialData.actionsToTake || "");
      setTreatmentPlan(initialData.treatment_plan || initialData.treatmentPlan || "");
      if (initialData.next_appointment_date || initialData.nextAppointmentDate) {
        const ms = initialData.next_appointment_date || initialData.nextAppointmentDate;
        if (ms) {
          const d = new Date(ms);
          setNextAppointmentDate(d.toISOString().split("T")[0]);
        }
      } else {
        setNextAppointmentDate("");
      }
      setNextAppointmentNote(initialData.next_appointment_note || initialData.nextAppointmentNote || "");
      setExpertNotes(initialData.expert_notes || initialData.expertNotes || "");
    } else {
      setDiagnosis("");
      setSymptoms("");
      setActionsToAvoid("");
      setActionsToTake("");
      setTreatmentPlan("");
      setNextAppointmentDate("");
      setNextAppointmentNote("");
      setExpertNotes("");
    }
  }, [initialData, open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!diagnosis.trim()) {
      showError("Vui lòng nhập chẩn đoán / tình trạng bệnh.");
      return;
    }

    const payload: SaveMedicalRecordRequest = {
      diagnosis: diagnosis.trim(),
      symptoms: symptoms.trim(),
      actions_to_avoid: actionsToAvoid.trim(),
      actions_to_take: actionsToTake.trim(),
      treatment_plan: treatmentPlan.trim(),
      next_appointment_date: nextAppointmentDate ? new Date(nextAppointmentDate).getTime() : null,
      next_appointment_note: nextAppointmentNote.trim(),
      expert_notes: expertNotes.trim(),
    };

    try {
      const result = await saveMutation.mutateAsync({
        appointmentId,
        payload,
      });
      showSuccess("Đã lưu hồ sơ bệnh án thành công!");
      onOpenChange(false);
      onSaved?.(result);
    } catch (err: any) {
      showError(err?.message || "Không thể lưu hồ sơ bệnh án. Vui lòng thử lại.");
    }
  };

  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Ghi chú hồ sơ bệnh án"
      description={`Tư vấn & chẩn đoán y khoa cho bệnh nhân: ${patientName || "Bệnh nhân"}`}
      size="2xl"
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Tình trạng bệnh / Chẩn đoán */}
        <div>
          <label htmlFor="mr-diagnosis" className="mb-1.5 flex items-center gap-1.5 text-sm font-medium text-foreground">
            <HeartPulse className="h-4 w-4 text-primary" />
            Tình trạng bệnh / Chẩn đoán <span className="text-danger">*</span>
          </label>
          <Input
            type="text"
            required
            id="mr-diagnosis"
            value={diagnosis}
            onChange={(e) => setDiagnosis(e.target.value)}
            placeholder="Ví dụ: Rối loạn lo âu lan tỏa, Căng thẳng mức độ nhẹ..."
          />
        </div>

        {/* Triệu chứng */}
        <div>
          <label htmlFor="mr-symptoms" className="mb-1.5 flex items-center gap-1.5 text-sm font-medium text-foreground">
            <AlertCircle className="h-4 w-4 text-warning" />
            Triệu chứng lâm sàng
          </label>
          <Textarea
            rows={2}
            id="mr-symptoms"
            value={symptoms}
            onChange={(e) => setSymptoms(e.target.value)}
            placeholder="Ví dụ: Mất ngủ thường xuyên về đêm, tim đập nhanh khi áp lực, khó tập trung..."
          />
        </div>

        {/* 2 cột: Cần tránh & Cần làm */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label htmlFor="mr-avoid" className="mb-1.5 flex items-center gap-1.5 text-sm font-medium text-danger">
              <ShieldAlert className="h-4 w-4" />
              Các hành động CẦN TRÁNH
            </label>
            <Textarea
              rows={3}
              id="mr-avoid"
            value={actionsToAvoid}
              onChange={(e) => setActionsToAvoid(e.target.value)}
              placeholder="Ví dụ: Tránh sử dụng chất kích thích (cà phê, rượu), tránh dùng điện thoại trước khi ngủ 1h..."
            />
          </div>

          <div>
            <label htmlFor="mr-take" className="mb-1.5 flex items-center gap-1.5 text-sm font-medium text-success">
              <CheckCircle2 className="h-4 w-4" />
              Các hành động CẦN LÀM / Lời dặn
            </label>
            <Textarea
              rows={3}
              id="mr-take"
            value={actionsToTake}
              onChange={(e) => setActionsToTake(e.target.value)}
              placeholder="Ví dụ: Tập thở 4-7-8 mỗi ngày 15 phút, đi bộ nhẹ nhàng buổi sáng, duy trì nhật ký cảm xúc..."
            />
          </div>
        </div>

        {/* Phác đồ / Kế hoạch điều trị */}
        <div>
          <label htmlFor="mr-treatment" className="mb-1.5 flex items-center gap-1.5 text-sm font-medium text-foreground">
            <FileText className="h-4 w-4 text-primary" />
            Phác đồ / Hướng can thiệp
          </label>
          <Textarea
            rows={2}
            id="mr-treatment"
            value={treatmentPlan}
            onChange={(e) => setTreatmentPlan(e.target.value)}
            placeholder="Ví dụ: Liệu pháp nhận thức hành vi (CBT), bài tập thư giãn cơ tiến triển..."
          />
        </div>

        {/* Buổi gặp tiếp theo */}
        <div className="rounded-xl border border-border bg-surface p-4 space-y-3">
          <p className="flex items-center gap-1.5 text-sm font-medium text-foreground">
            <Calendar className="h-4 w-4 text-primary" />
            Lịch hẹn tái khám / Buổi gặp tiếp theo (nếu có)
          </p>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="mr-next-date" className="mb-1 block text-xs font-medium text-muted-foreground">Ngày đề xuất</label>
              <Input
                type="date"
                id="mr-next-date"
            value={nextAppointmentDate}
                onChange={(e) => setNextAppointmentDate(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="mr-next-note" className="mb-1 block text-xs font-medium text-muted-foreground">Ghi chú lịch hẹn</label>
              <Input
                type="text"
                id="mr-next-note"
            value={nextAppointmentNote}
                onChange={(e) => setNextAppointmentNote(e.target.value)}
                placeholder="Ví dụ: Tái khám sau 2 tuần để đánh giá tiến triển..."
              />
            </div>
          </div>
        </div>

        {/* Ghi chú thêm */}
        <div>
          <label htmlFor="mr-notes" className="mb-1.5 flex items-center gap-1.5 text-sm font-medium text-muted-foreground">
            <Sparkles className="h-4 w-4" />
            Ghi chú bổ sung của Chuyên gia
          </label>
          <Textarea
            rows={2}
            id="mr-notes"
            value={expertNotes}
            onChange={(e) => setExpertNotes(e.target.value)}
            placeholder="Ghi chú thêm về phản ứng hoặc tâm lý bệnh nhân..."
          />
        </div>

        {/* Nút hành động */}
        <div className="flex flex-col-reverse gap-2 border-t border-border pt-4 sm:flex-row sm:justify-end sm:gap-3">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saveMutation.isPending}
          >
            Hủy
          </Button>
          <Button type="submit" disabled={saveMutation.isPending}>
            {saveMutation.isPending ? "Đang lưu..." : "Lưu hồ sơ bệnh án"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
