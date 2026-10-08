"use client";

import { useState } from "react";
import { Badge, Card, PageHeader, Pagination, Spinner } from "@/components/ui";
import {
  FulfillmentBadge,
  OrderStatusBadge,
  OrderTypeLabel,
  PaymentFilterBar,
  StatCard,
  defaultDateRange,
  formatDateTime,
  formatVnd,
  shortId,
} from "@/components/payment";
import { useApiQuery, useTranslation } from "@/hooks";
import { paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { PaymentOrderFilters } from "@/types";

const PAGE_SIZE = 10;

/** Expert earnings: every order paid to me with gross, platform commission and net amount. */
export default function ExpertWalletPage() {
  const { t } = useTranslation();
  const [filters, setFilters] = useState<PaymentOrderFilters>(defaultDateRange);
  const [page, setPage] = useState(0);

  const listParams = { ...filters, page, size: PAGE_SIZE };
  const { data: orders, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.expertPaymentOrders(listParams),
    queryFn: () => paymentApi.listExpertPaymentOrders(listParams),
  });
  const { data: summary } = useApiQuery({
    queryKey: QUERY_KEYS.expertPaymentSummary(filters),
    queryFn: () => paymentApi.getExpertPaymentSummary(filters),
  });

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        className="mb-0"
        title={t("payment.expert.title", "Earnings")}
        description={t("payment.expert.description", "Payments for your sessions, platform commission and what you receive.")}
      />

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <StatCard
          label={t("payment.summary.gross", "Gross revenue")}
          value={formatVnd(summary?.grossTotal ?? 0)}
        />
        <StatCard
          label={t("payment.summary.commission", "Commission")}
          value={formatVnd(summary?.commissionTotal ?? 0)}
          tone="danger"
        />
        <StatCard
          label={t("payment.summary.net", "Net amount")}
          value={formatVnd(summary?.netTotal ?? 0)}
          tone="success"
        />
        <StatCard
          label={t("payment.summary.orders", "Orders")}
          value={summary?.totalOrders ?? 0}
          hint={t("payment.summary.successOrders", "{{count}} paid", { count: summary?.successOrders ?? 0 })}
        />
      </div>

      <Card className="p-4">
        <PaymentFilterBar
          value={filters}
          onChange={(next) => {
            setFilters(next);
            setPage(0);
          }}
        />
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
                  <th className="px-4 py-3">{t("payment.col.date", "Date")}</th>
                  <th className="px-4 py-3">{t("payment.col.type", "Type")}</th>
                  <th className="px-4 py-3">{t("payment.col.patient", "Patient")}</th>
                  <th className="px-4 py-3 text-right">{t("payment.col.gross", "Gross")}</th>
                  <th className="px-4 py-3 text-right">{t("payment.col.commission", "Commission")}</th>
                  <th className="px-4 py-3 text-right">{t("payment.col.net", "Net")}</th>
                  <th className="px-4 py-3">{t("payment.col.status", "Status")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {orders?.items.map((order) => (
                  <tr key={order.id} className="transition-colors hover:bg-surface/60">
                    <td className="whitespace-nowrap px-4 py-3 text-muted-foreground">
                      {formatDateTime(order.paidAt ?? order.createdAt)}
                    </td>
                    <td className="px-4 py-3 text-foreground">
                      <OrderTypeLabel type={order.type} />
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground" title={order.payerId}>
                      {shortId(order.payerId)}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-right tabular-nums">{formatVnd(order.grossAmount)}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-right tabular-nums text-danger">
                      −{formatVnd(order.commissionAmount)}
                      <div className="text-xs text-muted-foreground">{Math.round(order.commissionRate * 100)}%</div>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-right font-semibold tabular-nums text-success">
                      {formatVnd(order.netAmount)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1.5">
                        <OrderStatusBadge status={order.status} />
                        {order.status === "SUCCESS" && (
                          <>
                            <FulfillmentBadge status={order.fulfillmentStatus} />
                            <Badge tone={order.released ? "success" : "neutral"}>
                              {order.released
                                ? t("payment.expert.released", "Available")
                                : t("payment.expert.onHold", "On hold")}
                            </Badge>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
                {orders?.items.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-6 py-10 text-center text-muted-foreground">
                      {t("payment.empty", "No transactions in this period.")}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          {orders && (
            <Pagination page={orders.page + 1} totalPages={orders.totalPages} onPageChange={(next) => setPage(next - 1)} />
          )}
        </Card>
      )}
    </div>
  );
}
