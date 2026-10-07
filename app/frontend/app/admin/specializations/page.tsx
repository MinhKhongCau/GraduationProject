"use client";

import { useState } from "react";
import { Search, Edit2, Plus, Power, Trash2 } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { Card, Button, Spinner, Pagination, PageHeader, Input, Select, Badge } from "@/components/ui";
import { useApiQuery, useApiMutation, useDebounce } from "@/hooks";
import { specializationApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import { SpecializationFormModal } from "./component/SpecializationFormModal";
import { DeleteSpecializationModal } from "./component/DeleteSpecializationModal";
import type { AdminSpecialization, CreateSpecializationRequest } from "@/types";

const PAGE_SIZE = 10;

type StatusFilter = "all" | "active" | "inactive";

const STATUS_FILTER_VALUES: Record<StatusFilter, boolean | undefined> = {
  all: undefined,
  active: true,
  inactive: false,
};

export default function AdminSpecializationsPage() {
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const debouncedQuery = useDebounce(query);
  /** undefined = modal closed, null = creating, otherwise editing that specialization. */
  const [editing, setEditing] = useState<AdminSpecialization | null | undefined>(undefined);
  const [deleting, setDeleting] = useState<AdminSpecialization | null>(null);

  const params = {
    page,
    pageSize: PAGE_SIZE,
    search: debouncedQuery || undefined,
    isActive: STATUS_FILTER_VALUES[statusFilter],
  };
  const { data, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.adminSpecializations(params),
    queryFn: () => specializationApi.listAllSpecializations(params),
  });

  /** Also refreshes the public active-only list used by booking, expert settings and admin expert edit. */
  function invalidateSpecializations() {
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.specializations() });
  }

  const saveMutation = useApiMutation({
    mutationFn: (values: CreateSpecializationRequest) =>
      editing ? specializationApi.updateSpecialization(editing.specId, values) : specializationApi.createSpecialization(values),
    onSuccess: () => {
      showSuccess(editing ? "Cập nhật chuyên khoa thành công!" : "Thêm chuyên khoa thành công!");
      invalidateSpecializations();
      setEditing(undefined);
    },
  });

  const statusMutation = useApiMutation({
    mutationFn: ({ specId, isActive }: { specId: string; isActive: boolean }) =>
      specializationApi.updateSpecializationStatus(specId, isActive),
    onSuccess: (spec) => {
      showSuccess(spec.isActive ? "Đã kích hoạt chuyên khoa." : "Đã ngừng hoạt động chuyên khoa.");
      invalidateSpecializations();
      setDeleting(null);
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: (specId: string) => specializationApi.deleteSpecialization(specId),
    onSuccess: () => {
      showSuccess("Đã xoá chuyên khoa.");
      invalidateSpecializations();
      setDeleting(null);
    },
  });

  const specializations = data?.items ?? [];

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        className="mb-0"
        title="Quản lý chuyên khoa"
        description="Các chuyên khoa chuyên gia có thể đăng ký và bệnh nhân chọn khi đặt lịch."
        actions={
          <Button onClick={() => setEditing(null)}>
            <Plus className="h-4 w-4" /> Thêm chuyên khoa
          </Button>
        }
      />

      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="relative w-full sm:w-72">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            aria-label="Tìm theo tên hoặc mã"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setPage(1);
            }}
            placeholder="Tìm theo tên hoặc mã..."
            className="pl-9"
          />
        </div>
        <Select
          aria-label="Lọc theo trạng thái"
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value as StatusFilter);
            setPage(1);
          }}
          className="sm:w-48"
        >
          <option value="all">Tất cả trạng thái</option>
          <option value="active">Đang hoạt động</option>
          <option value="inactive">Ngừng hoạt động</option>
        </Select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left text-sm">
              <thead className="border-b border-border bg-surface text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="px-4 py-3">Mã</th>
                  <th className="px-4 py-3">Tên chuyên khoa</th>
                  <th className="px-4 py-3">Slug</th>
                  <th className="px-4 py-3 text-right">Chuyên gia</th>
                  <th className="px-4 py-3">Trạng thái</th>
                  <th className="px-4 py-3 text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {specializations.map((spec) => (
                  <tr key={spec.specId} className="transition-colors hover:bg-surface/60">
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground">{spec.code}</td>
                    <td className="max-w-xs px-4 py-3">
                      <div className="font-medium text-foreground">{spec.name}</div>
                      {spec.symptoms.length > 0 && (
                        <div className="truncate text-xs text-muted-foreground">{spec.symptoms.join(", ")}</div>
                      )}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{spec.slug}</td>
                    <td className="px-4 py-3 text-right tabular-nums text-foreground">{spec.expertCount}</td>
                    <td className="px-4 py-3">
                      <Badge tone={spec.isActive ? "success" : "neutral"}>
                        {spec.isActive ? "Đang hoạt động" : "Ngừng hoạt động"}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1.5">
                        <Button
                          variant="ghost"
                          size="sm"
                          disabled={statusMutation.isPending}
                          onClick={() => statusMutation.mutate({ specId: spec.specId, isActive: !spec.isActive })}
                          className={spec.isActive ? "text-warning hover:bg-warning-soft" : "text-success hover:bg-success-soft"}
                        >
                          <Power className="h-3.5 w-3.5" /> {spec.isActive ? "Tắt" : "Bật"}
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditing(spec)}
                          className="text-primary hover:bg-primary-soft"
                        >
                          <Edit2 className="h-3.5 w-3.5" /> Sửa
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setDeleting(spec)}
                          className="text-danger hover:bg-danger-soft"
                          title={spec.expertCount > 0 ? "Đang có chuyên gia sử dụng, chỉ có thể ngừng hoạt động" : undefined}
                        >
                          <Trash2 className="h-3.5 w-3.5" /> Xoá
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
                {specializations.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-10 text-center text-muted-foreground">
                      Không tìm thấy chuyên khoa nào.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          {data && <Pagination page={data.page} totalPages={data.totalPages} onPageChange={setPage} />}
        </Card>
      )}

      <SpecializationFormModal
        open={editing !== undefined}
        onOpenChange={(open) => {
          if (!open) setEditing(undefined);
        }}
        specialization={editing ?? null}
        isPending={saveMutation.isPending}
        onSubmit={(values) => saveMutation.mutate(values)}
      />

      <DeleteSpecializationModal
        specialization={deleting}
        onOpenChange={(open) => {
          if (!open) setDeleting(null);
        }}
        isPending={deleteMutation.isPending || statusMutation.isPending}
        onConfirm={() => deleting && deleteMutation.mutate(deleting.specId)}
        onDeactivate={() => deleting && statusMutation.mutate({ specId: deleting.specId, isActive: false })}
      />
    </div>
  );
}
