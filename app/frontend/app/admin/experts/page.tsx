"use client";

import { useState } from "react";
import { Search, Edit2, BadgeCheck, ShieldX } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { Card, Button, Spinner, Pagination } from "@/components/ui";
import { useApiQuery, useApiMutation, useDebounce } from "@/hooks";
import { expertApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import { ExpertEditModal } from "./component/ExpertEditModal";
import type { ExpertVerificationStatus, UpdateExpertProfileRequest } from "@/types";

const PAGE_SIZE = 10;

const STATUS_STYLES: Record<ExpertVerificationStatus, string> = {
  VERIFIED: "bg-success-soft text-success",
  PENDING: "bg-warning-soft text-warning",
  UNVERIFIED: "bg-surface text-muted-foreground",
  REJECTED: "bg-danger-soft text-danger",
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

  const params = { page, pageSize: PAGE_SIZE, search: debouncedQuery || undefined };
  const { data, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.experts(params),
    queryFn: () => expertApi.listExperts(params),
  });

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
        variables.status === "VERIFIED" ? "Đã xác minh chuyên gia!" : "Đã từ chối hồ sơ chuyên gia."
      );
      invalidateExperts(variables.accountId);
    },
  });

  const experts = data?.items ?? [];

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Quản lý chuyên gia</h1>
          <p className="text-sm text-muted-foreground">
            Danh sách chuyên gia, chuyên khoa và trạng thái xác minh hồ sơ.
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
                  <th className="px-6 py-4">Chuyên khoa</th>
                  <th className="px-6 py-4">Trạng thái</th>
                  <th className="px-6 py-4 text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {experts.map((expert) => (
                  <tr key={expert.expertId} className="hover:bg-surface/30">
                    <td className="px-6 py-4 font-medium text-foreground">{expert.fullName}</td>
                    <td className="px-6 py-4 text-muted-foreground">
                      <div>{expert.email || "—"}</div>
                      <div className="text-xs">{expert.phoneNumber || "—"}</div>
                    </td>
                    <td className="max-w-xs px-6 py-4 text-muted-foreground">
                      {expert.specializations.map((spec) => spec.name).join(", ") || "—"}
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium ${STATUS_STYLES[expert.verificationStatus]}`}
                      >
                        {STATUS_LABELS[expert.verificationStatus]}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-1.5">
                        {expert.verificationStatus !== "VERIFIED" && (
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={verifyMutation.isPending}
                            onClick={() =>
                              verifyMutation.mutate({ accountId: expert.accountId, status: "VERIFIED" })
                            }
                            className="flex items-center gap-1 text-success hover:bg-success-soft/50"
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
                            className="flex items-center gap-1 text-danger hover:bg-danger-soft/50"
                          >
                            <ShieldX className="h-3.5 w-3.5" /> Từ chối
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditAccountId(expert.accountId)}
                          className="flex items-center gap-1 text-primary hover:bg-primary-soft/50"
                        >
                          <Edit2 className="h-3.5 w-3.5" /> Sửa
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
                {experts.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-6 py-10 text-center text-muted-foreground">
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
