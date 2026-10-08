"use client";

import { useState } from "react";
import { CalendarCheck, Wallet } from "lucide-react";
import { Card, Pagination, Spinner } from "@/components/ui";
import {
  OrderStatusBadge,
  OrderTypeLabel,
  PaymentFilterBar,
  StatCard,
  defaultDateRange,
  formatDateTime,
  formatVnd,
} from "@/components/payment";
import { useApiQuery, useTranslation } from "@/hooks";
import { paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { PaymentOrderFilters } from "@/types";

const PAGE_SIZE = 10;

/** Patient's own payments with total paid for the selected window. */
export function PaymentHistory() {
  const { t } = useTranslation();
  const [filters, setFilters] = useState<PaymentOrderFilters>(defaultDateRange);
  const [page, setPage] = useState(0);

  const listParams = { ...filters, page, size: PAGE_SIZE };
  const { data: orders, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.myPaymentOrders(listParams),
    queryFn: () => paymentApi.listMyPaymentOrders(listParams),
  });
  const { data: summary } = useApiQuery({
    queryKey: QUERY_KEYS.myPaymentSummary(filters),
    queryFn: () => paymentApi.getMyPaymentSummary(filters),
  });

  return (
    <Card className="mt-6 overflow-hidden">
      <div className="space-y-4 border-b border-border px-5 py-4">
        <h2 className="text-base font-semibold text-foreground">{t("payment.patient.history", "Payment history")}</h2>
        <div className="grid grid-cols-2 gap-3">
          <StatCard
            label={t("payment.patient.totalPaid", "Total paid")}
            value={formatVnd(summary?.totalPaid ?? 0)}
            tone="primary"
          />
          <StatCard
            label={t("payment.summary.orders", "Orders")}
            value={summary?.totalOrders ?? 0}
            hint={t("payment.summary.successOrders", "{{count}} paid", { count: summary?.successOrders ?? 0 })}
          />
        </div>
        <PaymentFilterBar
          value={filters}
          onChange={(next) => {
            setFilters(next);
            setPage(0);
          }}
        />
      </div>

      {isLoading ? (
        <div className="flex justify-center py-10">
          <Spinner className="h-6 w-6" />
        </div>
      ) : !orders || orders.items.length === 0 ? (
        <p className="px-5 py-10 text-center text-sm text-muted-foreground">
          {t("payment.empty", "No transactions in this period.")}
        </p>
      ) : (
        <div className="divide-y divide-border">
          {orders.items.map((order) => {
            const Icon = order.type === "TOP_UP" ? Wallet : CalendarCheck;
            return (
              <div key={order.id} className="flex items-center justify-between gap-4 px-5 py-4">
                <div className="flex min-w-0 items-center gap-4">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary">
                    <Icon className="h-5 w-5" aria-hidden="true" />
                  </div>
                  <div className="min-w-0">
                    <p className="text-sm font-semibold text-foreground">
                      <OrderTypeLabel type={order.type} />
                    </p>
                    <p className="mt-0.5 text-xs text-muted-foreground">
                      {formatDateTime(order.paidAt ?? order.createdAt)} · {order.gateway}
                    </p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-bold tabular-nums text-foreground">{formatVnd(order.amountVnd)}</p>
                  <div className="mt-1">
                    <OrderStatusBadge status={order.status} />
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
      {orders && (
        <Pagination page={orders.page + 1} totalPages={orders.totalPages} onPageChange={(next) => setPage(next - 1)} />
      )}
    </Card>
  );
}
