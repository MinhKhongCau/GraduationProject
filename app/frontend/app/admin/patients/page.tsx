"use client";

import { useState } from "react";
import { Search, Edit2 } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { Card, Button, Spinner, Pagination } from "@/components/ui";
import { useApiQuery, useApiMutation, useDebounce } from "@/hooks";
import { patientApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import { PatientEditModal } from "./component/PatientEditModal";
import type { UpdatePatientProfileRequest } from "@/types";

const PAGE_SIZE = 10;

export default function AdminPatientsPage() {
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState("");
  const debouncedQuery = useDebounce(query);
  const [editAccountId, setEditAccountId] = useState<string | null>(null);

  const params = { page, pageSize: PAGE_SIZE, search: debouncedQuery || undefined };
  const { data, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.patients(params),
    queryFn: () => patientApi.listPatients(params),
  });

  const updateMutation = useApiMutation({
    mutationFn: (payload: UpdatePatientProfileRequest) =>
      patientApi.updatePatientProfile(editAccountId!, payload),
    onSuccess: () => {
      showSuccess("Cập nhật hồ sơ bệnh nhân thành công!");
      queryClient.invalidateQueries({ queryKey: ["patients"] });
      if (editAccountId) {
        queryClient.invalidateQueries({ queryKey: QUERY_KEYS.patientProfile(editAccountId) });
      }
      setEditAccountId(null);
    },
  });

  const patients = data?.items ?? [];

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Quản lý bệnh nhân</h1>
          <p className="text-sm text-muted-foreground">
            Danh sách bệnh nhân đã đăng ký và hồ sơ chi tiết.
          </p>
        </div>
        <div className="relative w-full sm:w-72">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setPage(1);
            }}
            placeholder="Tìm theo tên..."
            className="w-full rounded-xl border border-border bg-background py-2 pl-9 pr-3 text-sm text-foreground outline-none focus:border-primary"
          />
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left text-sm">
              <thead className="border-b border-border bg-surface text-xs font-semibold uppercase text-muted-foreground">
                <tr>
                  <th className="px-6 py-4">Họ và tên</th>
                  <th className="px-6 py-4">Liên hệ</th>
                  <th className="px-6 py-4">Ngày sinh</th>
                  <th className="px-6 py-4">Giới tính</th>
                  <th className="px-6 py-4">Địa chỉ</th>
                  <th className="px-6 py-4 text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {patients.map((patient) => (
                  <tr key={patient.patientId} className="hover:bg-surface/30">
                    <td className="px-6 py-4 font-medium text-foreground">{patient.fullName}</td>
                    <td className="px-6 py-4 text-muted-foreground">
                      <div>{patient.email || "—"}</div>
                      <div className="text-xs">{patient.phoneNumber || "—"}</div>
                    </td>
                    <td className="px-6 py-4 text-muted-foreground">
                      {patient.dateOfBirth ? patient.dateOfBirth.slice(0, 10) : "—"}
                    </td>
                    <td className="px-6 py-4 text-muted-foreground">{patient.gender || "—"}</td>
                    <td className="max-w-xs truncate px-6 py-4 text-muted-foreground">
                      {patient.address || "—"}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditAccountId(patient.accountId)}
                          className="flex items-center gap-1 text-primary hover:bg-primary-soft/50"
                        >
                          <Edit2 className="h-3.5 w-3.5" /> Xem / Sửa
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
                {patients.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-10 text-center text-muted-foreground">
                      Không tìm thấy bệnh nhân nào.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          {data && <Pagination page={data.page} totalPages={data.totalPages} onPageChange={setPage} />}
        </Card>
      )}

      <PatientEditModal
        open={editAccountId !== null}
        onOpenChange={(open) => {
          if (!open) setEditAccountId(null);
        }}
        accountId={editAccountId}
        isPending={updateMutation.isPending}
        onSubmit={(values) => updateMutation.mutate(values)}
      />
    </div>
  );
}
