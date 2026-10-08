"use client";

import { useState } from "react";
import { CheckCircle2 } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { Button, Card, Input, Label, Modal, Pagination, Select, Spinner, Textarea } from "@/components/ui";
import { CompensationStatusBadge, defaultDateRange, formatDateTime, formatVnd } from "@/components/payment";
import { useApiMutation, useApiQuery } from "@/hooks";
import { paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import { ExpertSelect } from "./ExpertSelect";
import type { CompensationCase, CompensationCaseFilters, CompensationStatus, ExpertProfile } from "@/types";

const PAGE_SIZE = 10;

export interface CompensationCasesTabProps {
  experts: ExpertProfile[];
  expertName: (expertId: string) => string;
}

/** Orders that need manual review or a refund; the managing admin closes them once handled. */
export function CompensationCasesTab({ experts, expertName }: CompensationCasesTabProps) {
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();
  const [filters, setFilters] = useState<CompensationCaseFilters>(defaultDateRange);
  const [page, setPage] = useState(0);
  const [resolving, setResolving] = useState<CompensationCase | null>(null);
  const [note, setNote] = useState("");

  const updateFilters = (next: CompensationCaseFilters) => {
    setFilters(next);
    setPage(0);
  };

  const params = { ...filters, page, size: PAGE_SIZE };
  const { data, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.compensationCases(params),
    queryFn: () => paymentApi.listCompensationCases(params),
  });

  const resolveMutation = useApiMutation({
    mutationFn: () => paymentApi.resolveCompensationCase(resolving!.id, note || undefined),
    onSuccess: () => {
      showSuccess("Đã đóng hồ sơ xử lý.");
      setResolving(null);
      setNote("");
      queryClient.invalidateQueries({ queryKey: ["payment"] });
    },
  });

  return (
    <div className="space-y-6">
      <Card className="p-4">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <Label htmlFor="cases-status">Trạng thái</Label>
            <Select
              id="cases-status"
              value={filters.status ?? ""}
              onChange={(e) =>
                updateFilters({ ...filters, status: (e.target.value || undefined) as CompensationStatus | undefined })
              }
            >
              <option value="">Tất cả</option>
              <option value="MANUAL_REVIEW">Cần kiểm tra</option>
              <option value="REFUND_REQUIRED">Cần hoàn tiền</option>
              <option value="RESOLVED">Đã xử lý</option>
            </Select>
          </div>
          <ExpertSelect
            experts={experts}
            value={filters.expertId}
            onChange={(expertId) => updateFilters({ ...filters, expertId })}
          />
          <div>
            <Label htmlFor="cases-from">Từ ngày</Label>
            <Input
              id="cases-from"
              type="date"
              value={filters.from ?? ""}
              max={filters.to}
              onChange={(e) => updateFilters({ ...filters, from: e.target.value || undefined })}
            />
          </div>
          <div>
            <Label htmlFor="cases-to">Đến ngày</Label>
            <Input
              id="cases-to"
              type="date"
              value={filters.to ?? ""}
              min={filters.from}
              onChange={(e) => updateFilters({ ...filters, to: e.target.value || undefined })}
            />
          </div>
        </div>
      </Card>

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
                  <th className="px-4 py-3">Tạo lúc</th>
                  <th className="px-4 py-3">Chuyên gia</th>
                  <th className="px-4 py-3">Lý do</th>
                  <th className="px-4 py-3 text-right">Số tiền</th>
                  <th className="px-4 py-3">Trạng thái</th>
                  <th className="px-4 py-3 text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {data?.items.map((item) => (
                  <tr key={item.id} className="transition-colors hover:bg-surface/60">
                    <td className="whitespace-nowrap px-4 py-3 text-muted-foreground">{formatDateTime(item.createdAt)}</td>
                    <td className="px-4 py-3 font-medium text-foreground">{expertName(item.expertId)}</td>
                    <td className="max-w-sm px-4 py-3 text-muted-foreground">
                      <div className="text-xs font-semibold text-foreground">{item.reasonCode}</div>
                      <div>{item.safeReason}</div>
                      {item.resolutionNote && <div className="mt-1 text-xs italic">Ghi chú: {item.resolutionNote}</div>}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-right tabular-nums">{formatVnd(item.amountVnd)}</td>
                    <td className="px-4 py-3">
                      <CompensationStatusBadge status={item.status} />
                      {item.resolvedAt && (
                        <div className="mt-1 text-xs text-muted-foreground">{formatDateTime(item.resolvedAt)}</div>
                      )}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {item.status !== "RESOLVED" && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setResolving(item)}
                          className="text-success hover:bg-success-soft"
                        >
                          <CheckCircle2 className="h-3.5 w-3.5" /> Đóng hồ sơ
                        </Button>
                      )}
                    </td>
                  </tr>
                ))}
                {data?.items.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-10 text-center text-muted-foreground">
                      Không có hồ sơ cần xử lý.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          {data && <Pagination page={data.page + 1} totalPages={data.totalPages} onPageChange={(next) => setPage(next - 1)} />}
        </Card>
      )}

      <Modal
        open={resolving !== null}
        onOpenChange={(open) => {
          if (!open) {
            setResolving(null);
            setNote("");
          }
        }}
        title="Đóng hồ sơ xử lý"
        description={resolving ? `${resolving.reasonCode} · ${formatVnd(resolving.amountVnd)}` : undefined}
      >
        <div className="space-y-4">
          <div>
            <Label htmlFor="resolve-note">Kết quả xử lý (tuỳ chọn)</Label>
            <Textarea
              id="resolve-note"
              rows={3}
              maxLength={500}
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="VD: Đã hoàn tiền thủ công cho bệnh nhân."
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setResolving(null)}>
              Huỷ
            </Button>
            <Button disabled={resolveMutation.isPending} onClick={() => resolveMutation.mutate()}>
              Xác nhận đã xử lý
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
