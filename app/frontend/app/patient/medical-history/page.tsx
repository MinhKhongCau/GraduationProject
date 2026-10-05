"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { PersonalInfoSection } from "./component/PersonalInfoSection";
import { MedicalHistoryList } from "./component/MedicalHistoryList";
import { ConsultationRecordList } from "./component/ConsultationRecordList";
import { AddMedicalHistoryDialog } from "./component/AddMedicalHistoryDialog";
import { Card, PageHeader, Spinner, buttonClasses } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { patientApi } from "@/api";
import { QUERY_KEYS, ROUTES } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";
import { useErrorContext } from "@/context/ErrorContext";
import type { CreateMedicalHistoryRequest } from "@/types";
import { FileText, UserCheck, Stethoscope } from "lucide-react";

export default function MedicalHistoryPage() {
  const { user } = useAuthContext();
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();
  const [activeTab, setActiveTab] = useState<"consultation" | "personal">("consultation");
  const [addOpen, setAddOpen] = useState(false);

  const {
    data: profile,
    isLoading,
    isError,
  } = useQuery({
    queryKey: QUERY_KEYS.myPatientProfile(),
    queryFn: () => patientApi.getMyProfile(),
    enabled: !!user,
    retry: 0,
  });

  const { data: histories = [] } = useApiQuery({
    queryKey: QUERY_KEYS.myMedicalHistories(),
    queryFn: () => patientApi.getMedicalHistories(),
    enabled: !!profile,
  });

  const addHistoryMutation = useApiMutation({
    mutationFn: (payload: CreateMedicalHistoryRequest) => patientApi.addMedicalHistory(payload),
    onSuccess: () => {
      showSuccess("Đã thêm thông tin tiền sử y tế thành công.");
      setAddOpen(false);
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.myMedicalHistories() });
    },
  });

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner className="h-6 w-6" />
      </div>
    );
  }

  if (isError || !profile) {
    return (
      <Card className="mx-auto max-w-2xl p-10 text-center">
        <p className="mb-4 text-sm text-muted-foreground">
          Vui lòng hoàn tất thông tin cá nhân trong mục Cài đặt trước khi xem hồ sơ bệnh án.
        </p>
        <Link href={ROUTES.PATIENT.SETTINGS} className={buttonClasses()}>
          Đi tới Cài đặt
        </Link>
      </Card>
    );
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <PageHeader
        title="Hồ Sơ Y Tế & Bệnh Án"
        description="Theo dõi các kết luận chẩn đoán, lời dặn của chuyên gia và thông tin tiền sử bệnh cá nhân."
        className="mb-0"
      />

      {/* Tabs Switcher */}
      <div role="tablist" className="flex max-w-md gap-1 rounded-lg border border-border bg-background p-1">
        <button
          type="button"
          onClick={() => setActiveTab("consultation")}
          role="tab"
          aria-selected={activeTab === "consultation"}
          className={`flex h-9 flex-1 items-center justify-center gap-2 rounded-md px-3 text-sm font-semibold transition-colors ${
            activeTab === "consultation"
              ? "bg-background text-primary shadow-card"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          <Stethoscope className="h-4 w-4" aria-hidden="true" />
          Bệnh án từ Chuyên gia
        </button>
        <button
          type="button"
          onClick={() => setActiveTab("personal")}
          role="tab"
          aria-selected={activeTab === "personal"}
          className={`flex h-9 flex-1 items-center justify-center gap-2 rounded-md px-3 text-sm font-semibold transition-colors ${
            activeTab === "personal"
              ? "bg-background text-primary shadow-card"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          <UserCheck className="h-4 w-4" aria-hidden="true" />
          Tiền sử y tế cá nhân
        </button>
      </div>

      {activeTab === "consultation" ? (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="flex items-center gap-2 text-base font-semibold text-foreground">
              <FileText className="h-4 w-4 text-primary" aria-hidden="true" />
              Danh sách kết luận khám & hồ sơ bệnh án
            </h2>
          </div>
          <ConsultationRecordList />
        </div>
      ) : (
        <div className="space-y-6">
          <div>
            <h2 className="mb-3 text-base font-semibold text-foreground">Thông tin cá nhân</h2>
            <PersonalInfoSection profile={profile} />
          </div>

          <MedicalHistoryList histories={histories} onAdd={() => setAddOpen(true)} />
        </div>
      )}

      <AddMedicalHistoryDialog
        open={addOpen}
        onOpenChange={setAddOpen}
        onSubmit={(payload) => addHistoryMutation.mutate(payload)}
        isSubmitting={addHistoryMutation.isPending}
      />
    </div>
  );
}
