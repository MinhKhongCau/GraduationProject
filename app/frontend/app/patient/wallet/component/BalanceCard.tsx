import { Wallet, PlusCircle, ArrowDownToLine } from "lucide-react";
import { Button } from "@/components/ui";

export interface BalanceCardProps {
  balance: number;
  onTopUp: () => void;
  onWithdraw: () => void;
}

export function BalanceCard({ balance, onTopUp, onWithdraw }: BalanceCardProps) {
  return (
    <div className="flex flex-col items-center rounded-3xl bg-primary-soft p-8 text-center">
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary-soft-text/10 text-primary-soft-text">
        <Wallet className="h-6 w-6" />
      </div>
      <p className="mb-2 text-sm font-semibold uppercase tracking-wider text-muted-foreground">
        Total available balance
      </p>
      <h1 className="mb-8 text-5xl font-extrabold text-foreground">
        {(balance ?? 0).toLocaleString("vi-VN")} <span className="text-3xl">đ</span>
      </h1>

      <div className="flex w-full max-w-md gap-4">
        <Button onClick={onTopUp} className="flex-1">
          <PlusCircle className="h-5 w-5" /> Top up
        </Button>
        <Button onClick={onWithdraw} variant="outline" className="flex-1">
          <ArrowDownToLine className="h-5 w-5" /> Withdraw
        </Button>
      </div>
    </div>
  );
}
