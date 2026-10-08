"use client";

import { useState, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Button, Label, Modal, Spinner, Textarea } from "@/components/ui";
import {
  FulfillmentBadge,
  OrderStatusBadge,
  OrderTypeLabel,
  formatDateTime,
  formatVnd,
} from "@/components/payment";
import { formatAppointmentTime } from "@/components/appointment";
import { useApiMutation, useApiQuery } from "@/hooks";
import { paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type { AdminReviewAction } from "@/types";

export interface OrderDetailModalProps {
  orderId: string | null;
  expertName: (expertId: string) => string;
  onClose: () => void;
}

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="grid grid-cols-3 gap-3 py-2 text-sm">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="col-span-2 break-all text-foreground">{children}</dd>
    </div>
  );
}

/** Full order detail; a paid appointment order can be sent to manual review or refund. */
export function OrderDetailModal({ orderId, expertName, onClose }: OrderDetailModalProps) {
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();
  const [note, setNote] = useState("");

  const { data: order, isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.adminPaymentOrder(orderId ?? ""),
    queryFn: () => paymentApi.getAdminPaymentOrder(orderId!),
    enabled: !!orderId,
  });

  const reviewMutation = useApiMutation({
    mutationFn: (action: AdminReviewAction) => paymentApi.reviewPaymentOrder(orderId!, { action, note: note || undefined }),
    onSuccess: (_data, action) => {
      showSuccess(action === "REFUND_REQUIRED" ? "Đã tạo yêu cầu hoàn tiền." : "Đã chuyển đơn sang kiểm tra thủ công.");
      setNote("");
      queryClient.invalidateQueries({ queryKey: ["payment"] });
    },
  });

  const canReview =
    order?.status === "SUCCESS" &&
    order.type === "APPOINTMENT" &&
    order.fulfillmentStatus !== "MANUAL_REVIEW" &&
    order.fulfillmentStatus !== "REFUND_REQUIRED";

  return (
    <Modal
      open={orderId !== null}
      onOpenChange={(open) => {
        if (!open) {
          setNote("");
          onClose();
        }
      }}
      title="Chi tiết giao dịch"
      description={orderId ?? undefined}
      size="2xl"
    >
      {isLoading || !order ? (
        <div className="flex justify-center py-10">
          <Spinner className="h-6 w-6" />
        </div>
      ) : (
        <div className="space-y-6">
          <dl className="divide-y divide-border">
            <Row label="Trạng thái">
              <div className="flex flex-wrap gap-1.5">
                <OrderStatusBadge status={order.status} />
                <FulfillmentBadge status={order.fulfillmentStatus} />
              </div>
            </Row>
            <Row label="Loại"><OrderTypeLabel type={order.type} /></Row>
            <Row label="Chuyên gia">{order.expert?.fullName ?? expertName(order.expertId)}</Row>
            <Row label="Người thanh toán">
              {order.payer?.fullName ?? "—"}
              <span className="block text-xs text-muted-foreground">{order.payer?.email ?? order.payerId}</span>
            </Row>
            <Row label="Lịch hẹn">
              {order.appointment ? (
                <>
                  {formatAppointmentTime(order.appointment.startTime, order.appointment.endTime)}
                  <span className="block text-xs text-muted-foreground">
                    {[order.appointment.specializationName, order.appointment.patientFullName && `Người khám: ${order.appointment.patientFullName}`, order.appointment.status]
                      .filter(Boolean)
                      .join(" · ")}
                  </span>
                </>
              ) : (
                order.appointmentId ?? "—"
              )}
            </Row>
            <Row label="Tổng tiền">{formatVnd(order.grossAmount)}</Row>
            <Row label="Hoa hồng">
              {formatVnd(order.commissionAmount)} ({Math.round(order.commissionRate * 100)}%)
            </Row>
            <Row label="Chuyên gia thực nhận">{formatVnd(order.netAmount)}</Row>
            <Row label="Đã giải ngân cho chuyên gia">{order.released ? "Có" : "Chưa (đang tạm giữ)"}</Row>
            <Row label="Cổng thanh toán">{order.gateway}</Row>
            <Row label="Mã giao dịch cổng">{order.gatewayTxnRef || "—"}</Row>
            <Row label="Mã phản hồi / trạng thái">
              {order.gatewayResponseCode || "—"} / {order.gatewayTransactionStatus || "—"}
            </Row>
            <Row label="Thu tiền tại cổng">{order.gatewayCaptureStatus}</Row>
            <Row label="Tạo lúc">{formatDateTime(order.createdAt)}</Row>
            <Row label="Hết hạn lúc">{formatDateTime(order.expiresAt)}</Row>
            <Row label="Thanh toán lúc">{formatDateTime(order.paidAt)}</Row>
          </dl>

          {canReview && (
            <div className="space-y-3 rounded-xl border border-border bg-surface p-4">
              <p className="text-sm font-semibold text-foreground">Xử lý giao dịch</p>
              <div>
                <Label htmlFor="review-note">Ghi chú (tuỳ chọn)</Label>
                <Textarea
                  id="review-note"
                  rows={3}
                  maxLength={500}
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  placeholder="Lý do kiểm tra hoặc hoàn tiền..."
                />
              </div>
              <div className="flex flex-wrap justify-end gap-2">
                <Button
                  variant="outline"
                  disabled={reviewMutation.isPending}
                  onClick={() => reviewMutation.mutate("MANUAL_REVIEW")}
                >
                  Kiểm tra thủ công
                </Button>
                <Button
                  variant="danger"
                  disabled={reviewMutation.isPending}
                  onClick={() => reviewMutation.mutate("REFUND_REQUIRED")}
                >
                  Yêu cầu hoàn tiền
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </Modal>
  );
}
