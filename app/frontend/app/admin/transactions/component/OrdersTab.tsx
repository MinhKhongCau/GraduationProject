"use client";

import { useState } from "react";
import { Eye } from "lucide-react";
import { Button, Card, Label, Pagination, Select, Spinner } from "@/components/ui";
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
import { useApiQuery } from "@/hooks";
import { paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { ExpertSelect } from "./ExpertSelect";
import { OrderDetailModal } from "./OrderDetailModal";
import type { AdminPaymentOrderFilters, ExpertProfile, FulfillmentStatus } from "@/types";

const PAGE_SIZE = 10;

export interface OrdersTabProps {
  experts: ExpertProfile[];
  expertName: (expertId: string) => string;
}

export function OrdersTab({ experts, expertName }: OrdersTabProps) {
  const [filters, setFilters] = useState<AdminPaymentOrderFilters>(defaultDateRange);
  const [page, setPage] = useState(0);
  const [detailId, setDetailId] = useState<string | null>(null);

  const updateFilters = (next: AdminPaymentOrderFilters) => {
    setFilters(next);
    setPage(0);
  };

  const listParams = { ...filters, page, size: PAGE_SIZE };
  const { data: orders, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.adminPaymentOrders(listParams),
    queryFn: () => paymentApi.listAdminPaymentOrders(listParams),
  });
  const { data: summary } = useApiQuery({
    queryKey: QUERY_KEYS.adminPaymentSummary(filters),
    queryFn: () => paymentApi.getAdminPaymentSummary(filters),
  });

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <StatCard label="Tổng doanh thu" value={formatVnd(summary?.grossTotal ?? 0)} hint="Chỉ tính đơn đã thanh toán" />
        <StatCard label="Hoa hồng nền tảng" value={formatVnd(summary?.commissionTotal ?? 0)} tone="primary" />
        <StatCard label="Chuyên gia thực nhận" value={formatVnd(summary?.netTotal ?? 0)} tone="success" />
        <StatCard
          label="Số đơn"
          value={summary?.totalOrders ?? 0}
          hint={`${summary?.successOrders ?? 0} thành công · ${summary?.pendingOrders ?? 0} chờ · ${summary?.failedOrders ?? 0} lỗi · ${summary?.expiredOrders ?? 0} hết hạn`}
        />
      </div>

      <Card className="p-4">
        <PaymentFilterBar value={filters} onChange={updateFilters}>
          <ExpertSelect
            experts={experts}
            value={filters.expertId}
            onChange={(expertId) => updateFilters({ ...filters, expertId })}
          />
          <div>
            <Label htmlFor="transactions-fulfillment">Xử lý lịch hẹn</Label>
            <Select
              id="transactions-fulfillment"
              value={filters.fulfillmentStatus ?? ""}
              onChange={(e) =>
                updateFilters({ ...filters, fulfillmentStatus: (e.target.value || undefined) as FulfillmentStatus | undefined })
              }
            >
              <option value="">Tất cả</option>
              <option value="BOOKING_CONFIRMED">Đã xác nhận lịch</option>
              <option value="PENDING">Đang xử lý</option>
              <option value="BOOKING_FAILED">Đặt lịch lỗi</option>
              <option value="MANUAL_REVIEW">Cần kiểm tra</option>
              <option value="REFUND_REQUIRED">Cần hoàn tiền</option>
            </Select>
          </div>
        </PaymentFilterBar>
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
                  <th className="px-4 py-3">Bệnh nhân</th>
                  <th className="px-4 py-3">Loại</th>
                  <th className="px-4 py-3 text-right">Tổng tiền</th>
                  <th className="px-4 py-3 text-right">Hoa hồng</th>
                  <th className="px-4 py-3 text-right">Thực nhận</th>
                  <th className="px-4 py-3">Trạng thái</th>
                  <th className="px-4 py-3 text-right">Chi tiết</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {orders?.items.map((order) => (
                  <tr key={order.id} className="transition-colors hover:bg-surface/60">
                    <td className="whitespace-nowrap px-4 py-3 text-muted-foreground">
                      {formatDateTime(order.paidAt ?? order.createdAt)}
                    </td>
                    <td className="px-4 py-3 font-medium text-foreground">{expertName(order.expertId)}</td>
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground" title={order.payerId}>
                      {shortId(order.payerId)}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">
                      <OrderTypeLabel type={order.type} />
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-right tabular-nums">{formatVnd(order.grossAmount)}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-right tabular-nums text-primary">
                      {formatVnd(order.commissionAmount)}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-right font-semibold tabular-nums text-success">
                      {formatVnd(order.netAmount)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1.5">
                        <OrderStatusBadge status={order.status} />
                        {order.status === "SUCCESS" && <FulfillmentBadge status={order.fulfillmentStatus} />}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setDetailId(order.id)}
                        className="text-primary hover:bg-primary-soft"
                      >
                        <Eye className="h-3.5 w-3.5" /> Xem
                      </Button>
                    </td>
                  </tr>
                ))}
                {orders?.items.length === 0 && (
                  <tr>
                    <td colSpan={9} className="px-6 py-10 text-center text-muted-foreground">
                      Không có giao dịch nào trong khoảng thời gian này.
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

      <OrderDetailModal orderId={detailId} expertName={expertName} onClose={() => setDetailId(null)} />
    </div>
  );
}
