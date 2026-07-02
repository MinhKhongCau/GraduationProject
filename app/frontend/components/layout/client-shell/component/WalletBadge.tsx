"use client";

import { Wallet } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useAuthContext } from "@/context/AuthContext";

export function WalletBadge() {
  const { user } = useAuthContext();

  // Plain useQuery (not the auto-toast useApiQuery) — a user without a
  // wallet yet (never called payments/wallets/init) shouldn't see an error
  // toast just for viewing the header; the wallet page itself surfaces
  // real errors when the user is actually managing their wallet.
  const { data: wallet } = useQuery({
    queryKey: user ? QUERY_KEYS.wallet(user.id) : ["wallet", "anonymous"],
    queryFn: () => paymentApi.getWallet(user!.id),
    enabled: !!user,
    retry: 0,
  });

  if (!user) return null;

  return (
    <div className="hidden items-center gap-1.5 rounded-full bg-surface px-3 py-1.5 text-xs font-semibold text-muted-foreground sm:flex">
      <Wallet className="h-3.5 w-3.5 text-primary" />
      <span className="text-foreground">
        {(wallet?.balance ?? 0).toLocaleString("vi-VN")}
        <span className="ml-0.5">đ</span>
      </span>
    </div>
  );
}
