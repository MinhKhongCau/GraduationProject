"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { PersonalInfoSection } from "./component/PersonalInfoSection";
import { MedicalHistoryList } from "./component/MedicalHistoryList";
import { ConsultationRecordList } from "./component/ConsultationRecordList";
import { AddMedicalHistoryDialog } from "./component/AddMedicalHistoryDialog";
import { Button, Spinner } from "@/components/ui";
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
      <div className="mx-auto max-w-2xl rounded-3xl bg-surface p-10 text-center">
        <p className="mb-4 text-sm text-muted-foreground">
          Vui lòng hoàn tất thông tin cá nhân trong mục Cài đặt trước khi xem hồ sơ bệnh án.
        </p>
        <Link href={ROUTES.PATIENT.SETTINGS}>
          <Button>Đi tới Cài đặt</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Hồ Sơ Y Tế & Bệnh Án</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Theo dõi các kết luận chẩn đoán, lời dặn của chuyên gia và thông tin tiền sử bệnh cá nhân.
        </p>
      </div>

      {/* Tabs Switcher */}
      <div className="flex rounded-2xl bg-surface p-1 max-w-md border border-border">
        <button
          type="button"
          onClick={() => setActiveTab("consultation")}
          className={`flex flex-1 items-center justify-center gap-2 rounded-xl py-2.5 text-xs font-bold transition-all ${
            activeTab === "consultation"
              ? "bg-background text-primary shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          <Stethoscope className="h-4 w-4" />
          Bệnh án từ Chuyên gia
        </button>
        <button
          type="button"
          onClick={() => setActiveTab("personal")}
          className={`flex flex-1 items-center justify-center gap-2 rounded-xl py-2.5 text-xs font-bold transition-all ${
            activeTab === "personal"
              ? "bg-background text-primary shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          <UserCheck className="h-4 w-4" />
          Tiền sử y tế cá nhân
        </button>
      </div>

      {activeTab === "consultation" ? (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
              <FileText className="h-4 w-4 text-primary" />
              Danh sách kết luận khám & hồ sơ bệnh án
            </h3>
          </div>
          <ConsultationRecordList />
        </div>
      ) : (
        <div className="space-y-6">
          <div>
            <h3 className="mb-4 text-sm font-bold text-foreground">Thông tin cá nhân</h3>
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
