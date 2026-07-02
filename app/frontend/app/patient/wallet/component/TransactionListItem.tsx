import { ArrowDownLeft, ArrowUpRight, Building, RotateCcw } from "lucide-react";
import type { Transaction, TransactionType } from "@/types";

const TYPE_STYLES: Record<TransactionType, { icon: typeof ArrowUpRight; bg: string; text: string }> = {
  TOP_UP: { icon: ArrowDownLeft, bg: "bg-primary-soft", text: "text-primary" },
  EARNING: { icon: ArrowDownLeft, bg: "bg-primary-soft", text: "text-primary" },
  PAYMENT: { icon: ArrowUpRight, bg: "bg-danger-soft", text: "text-danger" },
  COMMISSION: { icon: ArrowUpRight, bg: "bg-danger-soft", text: "text-danger" },
  WITHDRAW: { icon: Building, bg: "bg-surface", text: "text-muted-foreground" },
  REFUND_WITHDRAW: { icon: RotateCcw, bg: "bg-surface", text: "text-muted-foreground" },
};

const LABELS: Record<TransactionType, string> = {
  TOP_UP: "Top up",
  EARNING: "Earning",
  PAYMENT: "Session payment",
  COMMISSION: "Platform commission",
  WITHDRAW: "Withdrawal",
  REFUND_WITHDRAW: "Withdrawal refund",
};

export function TransactionListItem({ transaction }: { transaction: Transaction }) {
  const style = TYPE_STYLES[transaction.transactionType];
  const Icon = style.icon;
  const isPositive = transaction.amount > 0;

  return (
    <div className="flex items-center justify-between rounded-2xl p-4 transition-colors hover:bg-surface">
      <div className="flex items-center gap-4">
        <div className={`flex h-10 w-10 items-center justify-center rounded-full ${style.bg} ${style.text}`}>
          <Icon className="h-5 w-5" />
        </div>
        <div>
          <p className="text-sm font-bold text-foreground">{LABELS[transaction.transactionType]}</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            {new Date(transaction.createdAt).toLocaleString("en-GB")}
          </p>
        </div>
      </div>
      <div className="text-right">
        <p className={`text-sm font-bold ${isPositive ? "text-success" : "text-danger"}`}>
          {isPositive ? "+" : ""}
          {transaction.amount.toLocaleString("vi-VN")} đ
        </p>
        <span
          className={`mt-1 inline-block rounded-md px-2 py-0.5 text-[10px] font-bold ${
            transaction.status === "PENDING" ? "bg-warning-soft text-warning" : "bg-surface text-muted-foreground"
          }`}
        >
          {transaction.status}
        </span>
      </div>
    </div>
  );
}
