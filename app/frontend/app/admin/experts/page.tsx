"use client";

import { useState } from "react";
import { Search, Edit2, BadgeCheck, ShieldX } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { Card, Button, Spinner, Pagination, PageHeader, Input, Select, Badge, type BadgeTone } from "@/components/ui";
import { useApiQuery, useApiMutation, useDebounce } from "@/hooks";
import { expertApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import { ExpertEditModal } from "./component/ExpertEditModal";
import type { ExpertVerificationStatus, UpdateExpertProfileRequest } from "@/types";

const PAGE_SIZE = 10;

const STATUS_TONES: Record<ExpertVerificationStatus, BadgeTone> = {
  VERIFIED: "success",
  PENDING: "warning",
  UNVERIFIED: "neutral",
  REJECTED: "danger",
};

const STATUS_LABELS: Record<ExpertVerificationStatus, string> = {
  VERIFIED: "Đã xác minh",
  PENDING: "Chờ duyệt",
  UNVERIFIED: "Chưa xác minh",
  REJECTED: "Đã từ chối",
};

export default function AdminExpertsPage() {
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState("");
  const debouncedQuery = useDebounce(query);
  const [editAccountId, setEditAccountId] = useState<string | null>(null);
  // Admin duyệt chuyên gia sẽ trở thành người quản lý chuyên gia đó (giao dịch, hồ sơ xử lý...).
  const [onlyManaged, setOnlyManaged] = useState(false);

  const params = { page, pageSize: PAGE_SIZE, search: debouncedQuery || undefined };
  const { data, isLoading } = useApiQuery({
    queryKey: onlyManaged ? QUERY_KEYS.managedExperts(params) : QUERY_KEYS.experts(params),
    queryFn: () => (onlyManaged ? expertApi.listManagedExperts(params) : expertApi.listExperts(params)),
  });

  // Public expert list hides the manager, so mark rows using the admin's own managed list.
  const managedParams = { page: 1, pageSize: 100 };
  const { data: managed } = useApiQuery({
    queryKey: QUERY_KEYS.managedExperts(managedParams),
    queryFn: () => expertApi.listManagedExperts(managedParams),
  });
  const managedIds = new Set((managed?.items ?? []).map((expert) => expert.accountId));

  function invalidateExperts(accountId?: string) {
    queryClient.invalidateQueries({ queryKey: ["experts"] });
    if (accountId) {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.expertProfile(accountId) });
    }
  }

  const updateMutation = useApiMutation({
    mutationFn: (payload: UpdateExpertProfileRequest) =>
      expertApi.updateExpertProfile(editAccountId!, payload),
    onSuccess: () => {
      showSuccess("Cập nhật hồ sơ chuyên gia thành công!");
      invalidateExperts(editAccountId ?? undefined);
      setEditAccountId(null);
    },
  });

  const verifyMutation = useApiMutation({
    mutationFn: ({ accountId, status }: { accountId: string; status: ExpertVerificationStatus }) =>
      expertApi.updateExpertVerification(accountId, status),
    onSuccess: (_data, variables) => {
      showSuccess(
        variables.status === "VERIFIED"
          ? "Đã xác minh chuyên gia! Bạn đã trở thành người quản lý chuyên gia này."
          : "Đã từ chối hồ sơ chuyên gia."
      );
      invalidateExperts(variables.accountId);
    },
  });

  const experts = data?.items ?? [];

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        className="mb-0"
        title="Quản lý chuyên gia"
        description="Danh sách chuyên gia, chuyên khoa và trạng thái xác minh hồ sơ."
        actions={
          <>
          <Select
            aria-label="Phạm vi"
            value={onlyManaged ? "managed" : "all"}
            onChange={(e) => {
              setOnlyManaged(e.target.value === "managed");
              setPage(1);
            }}
            className="w-full sm:w-56"
          >
            <option value="all">Tất cả chuyên gia</option>
            <option value="managed">Chuyên gia tôi quản lý</option>
          </Select>
          <div className="relative w-full sm:w-72">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              aria-label="Tìm theo tên"
              value={query}
              onChange={(e) => {
                setQuery(e.target.value);
                setPage(1);
              }}
              placeholder="Tìm theo tên..."
              className="pl-9"
            />
          </div>
          </>
        }
      />

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
                  <th className="px-4 py-3">Họ và tên</th>
                  <th className="px-4 py-3">Liên hệ</th>
                  <th className="px-4 py-3">Chuyên khoa</th>
                  <th className="px-4 py-3">Trạng thái</th>
                  <th className="px-4 py-3">Quản lý</th>
                  <th className="px-4 py-3 text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {experts.map((expert) => (
                  <tr key={expert.expertId} className="transition-colors hover:bg-surface/60">
                    <td className="px-4 py-3 font-medium text-foreground">{expert.fullName}</td>
                    <td className="px-4 py-3 text-muted-foreground">
                      <div>{expert.email || "—"}</div>
                      <div className="text-xs">{expert.phoneNumber || "—"}</div>
                    </td>
                    <td className="max-w-xs px-4 py-3 text-muted-foreground">
                      {expert.specializations.map((spec) => spec.name).join(", ") || "—"}
                    </td>
                    <td className="px-4 py-3">
                      <Badge tone={STATUS_TONES[expert.verificationStatus]}>
                        {STATUS_LABELS[expert.verificationStatus]}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      {managedIds.has(expert.accountId) ? (
                        <Badge tone="primary">Bạn quản lý</Badge>
                      ) : (
                        <span className="text-muted-foreground">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1.5">
                        {expert.verificationStatus !== "VERIFIED" && (
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={verifyMutation.isPending}
                            onClick={() =>
                              verifyMutation.mutate({ accountId: expert.accountId, status: "VERIFIED" })
                            }
                            className="text-success hover:bg-success-soft"
                          >
                            <BadgeCheck className="h-3.5 w-3.5" /> Duyệt
                          </Button>
                        )}
                        {expert.verificationStatus !== "REJECTED" && (
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={verifyMutation.isPending}
                            onClick={() =>
                              verifyMutation.mutate({ accountId: expert.accountId, status: "REJECTED" })
                            }
                            className="text-danger hover:bg-danger-soft"
                          >
                            <ShieldX className="h-3.5 w-3.5" /> Từ chối
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditAccountId(expert.accountId)}
                          className="text-primary hover:bg-primary-soft"
                        >
                          <Edit2 className="h-3.5 w-3.5" /> Sửa
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
                {experts.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-10 text-center text-muted-foreground">
                      Không tìm thấy chuyên gia nào.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          {data && <Pagination page={data.page} totalPages={data.totalPages} onPageChange={setPage} />}
        </Card>
      )}

      <ExpertEditModal
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
