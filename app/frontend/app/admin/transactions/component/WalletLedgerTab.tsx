"use client";

import { useState } from "react";
import { Card, Input, Label, Pagination, Select, Spinner } from "@/components/ui";
import { defaultDateRange, formatDateTime, formatVnd } from "@/components/payment";
import { useApiQuery } from "@/hooks";
import { paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { ExpertSelect } from "./ExpertSelect";
import type { ExpertProfile, ManagedWalletTransactionFilters, WalletTransactionType } from "@/types";

const PAGE_SIZE = 10;

const TYPE_LABELS: Record<WalletTransactionType, string> = {
  PAYMENT_RECEIVED: "Nhận thanh toán",
  COMMISSION_DEDUCTED: "Trừ hoa hồng",
  REFUND: "Hoàn tiền",
  WITHDRAWAL_LOCKED: "Khoá tiền rút",
  WITHDRAWAL_COMPLETED: "Rút tiền thành công",
  WITHDRAWAL_REJECTED: "Rút tiền bị từ chối",
  ADJUSTMENT: "Điều chỉnh",
};

export interface WalletLedgerTabProps {
  experts: ExpertProfile[];
  expertName: (expertId: string) => string;
}

/** Wallet ledger rows of managed experts (credits > 0, debits < 0). */
export function WalletLedgerTab({ experts, expertName }: WalletLedgerTabProps) {
  const [filters, setFilters] = useState<ManagedWalletTransactionFilters>(defaultDateRange);
  const [page, setPage] = useState(0);

  const updateFilters = (next: ManagedWalletTransactionFilters) => {
    setFilters(next);
    setPage(0);
  };

  const params = { ...filters, page, size: PAGE_SIZE };
  const { data, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.managedWalletTransactions(params),
    queryFn: () => paymentApi.listManagedWalletTransactions(params),
  });

  return (
    <div className="space-y-6">
      <Card className="p-4">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-5">
          <ExpertSelect
            experts={experts}
            value={filters.expertId}
            onChange={(expertId) => updateFilters({ ...filters, expertId })}
          />
          <div>
            <Label htmlFor="ledger-type">Loại giao dịch</Label>
            <Select
              id="ledger-type"
              value={filters.type ?? ""}
              onChange={(e) =>
                updateFilters({ ...filters, type: (e.target.value || undefined) as WalletTransactionType | undefined })
              }
            >
              <option value="">Tất cả</option>
              {Object.entries(TYPE_LABELS).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <Label htmlFor="ledger-direction">Chiều</Label>
            <Select
              id="ledger-direction"
              value={filters.direction ?? ""}
              onChange={(e) =>
                updateFilters({ ...filters, direction: (e.target.value || undefined) as "CREDIT" | "DEBIT" | undefined })
              }
            >
              <option value="">Tất cả</option>
              <option value="CREDIT">Tiền vào</option>
              <option value="DEBIT">Tiền ra</option>
            </Select>
          </div>
          <div>
            <Label htmlFor="ledger-from">Từ ngày</Label>
            <Input
              id="ledger-from"
              type="date"
              value={filters.from ?? ""}
              max={filters.to}
              onChange={(e) => updateFilters({ ...filters, from: e.target.value || undefined })}
            />
          </div>
          <div>
            <Label htmlFor="ledger-to">Đến ngày</Label>
            <Input
              id="ledger-to"
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
                  <th className="px-4 py-3">Thời gian</th>
                  <th className="px-4 py-3">Chuyên gia</th>
                  <th className="px-4 py-3">Loại</th>
                  <th className="px-4 py-3">Tham chiếu</th>
                  <th className="px-4 py-3 text-right">Số tiền</th>
                  <th className="px-4 py-3 text-right">Số dư sau</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {data?.items.map((tx) => (
                  <tr key={tx.id} className="transition-colors hover:bg-surface/60">
                    <td className="whitespace-nowrap px-4 py-3 text-muted-foreground">{formatDateTime(tx.createdAt)}</td>
                    <td className="px-4 py-3 font-medium text-foreground">{expertName(tx.expertId)}</td>
                    <td className="px-4 py-3 text-foreground">{TYPE_LABELS[tx.type] ?? tx.type}</td>
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground" title={tx.referenceId}>
                      {tx.referenceType} · {tx.referenceId.slice(0, 8)}
                    </td>
                    <td
                      className={`whitespace-nowrap px-4 py-3 text-right font-semibold tabular-nums ${
                        tx.amount >= 0 ? "text-success" : "text-danger"
                      }`}
                    >
                      {tx.amount >= 0 ? "+" : ""}
                      {formatVnd(tx.amount)}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-right tabular-nums text-muted-foreground">
                      {formatVnd(tx.balanceAfter)}
                    </td>
                  </tr>
                ))}
                {data?.items.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-10 text-center text-muted-foreground">
                      Không có giao dịch ví nào trong khoảng thời gian này.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          {data && <Pagination page={data.page + 1} totalPages={data.totalPages} onPageChange={(next) => setPage(next - 1)} />}
        </Card>
      )}
    </div>
  );
}
