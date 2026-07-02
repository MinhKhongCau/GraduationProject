import Link from "next/link";
import { Wallet, ArrowRight } from "lucide-react";
import { Card } from "@/components/ui";
import { ROUTES } from "@/constants";

export function WalletSnapshotCard({ balance }: { balance: number | undefined }) {
  return (
    <Card className="p-6">
      <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-full bg-primary-soft text-primary-soft-text">
        <Wallet className="h-5 w-5" />
      </div>
      <p className="mb-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">Wallet balance</p>
      <h3 className="mb-4 text-2xl font-extrabold text-foreground">
        {(balance ?? 0).toLocaleString("vi-VN")} <span className="text-base font-semibold">đ</span>
      </h3>
      <Link href={ROUTES.PATIENT.WALLET} className="inline-flex items-center gap-1 text-sm font-semibold text-primary hover:underline">
        Manage wallet <ArrowRight className="h-3.5 w-3.5" />
      </Link>
    </Card>
  );
}
