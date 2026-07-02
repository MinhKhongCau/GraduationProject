"use client";

import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { BalanceCard } from "./component/BalanceCard";
import { TransactionListItem } from "./component/TransactionListItem";
import { TopUpDialog } from "./component/TopUpDialog";
import { WithdrawDialog } from "./component/WithdrawDialog";
import { Button, Spinner } from "@/components/ui";
import { useApiMutation } from "@/hooks";
import { paymentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { TRANSACTIONS_MOCK } from "@/data";
import { useAuthContext } from "@/context/AuthContext";
import { useErrorContext } from "@/context/ErrorContext";
import type { CreateWithdrawalRequest } from "@/types";

export default function WalletPage() {
  const { user } = useAuthContext();
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();
  const [topUpOpen, setTopUpOpen] = useState(false);
  const [withdrawOpen, setWithdrawOpen] = useState(false);

  // Plain useQuery: a 404 here means the wallet just hasn't been
  // initialized yet, which is an expected state, not an error to toast.
  const {
    data: wallet,
    isLoading,
    isError,
  } = useQuery({
    queryKey: user ? QUERY_KEYS.wallet(user.id) : ["wallet", "anonymous"],
    queryFn: () => paymentApi.getWallet(user!.id),
    enabled: !!user,
    retry: 0,
  });

  const initWalletMutation = useApiMutation({
    mutationFn: () => paymentApi.initWallet({ ownerId: user!.id, userType: "PATIENT" }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: QUERY_KEYS.wallet(user!.id) }),
  });

  const topUpMutation = useApiMutation({
    mutationFn: (amount: number) => paymentApi.topUpWallet(user!.id, { amount }),
    onSuccess: () => {
      showSuccess("Top up successful.");
      setTopUpOpen(false);
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.wallet(user!.id) });
    },
  });

  const withdrawMutation = useApiMutation({
    mutationFn: (payload: CreateWithdrawalRequest) => paymentApi.requestWithdrawal(user!.id, payload),
    onSuccess: () => {
      showSuccess("Withdrawal request submitted.");
      setWithdrawOpen(false);
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.wallet(user!.id) });
    },
  });

  return (
    <div className="mx-auto max-w-4xl">
      <h1 className="mb-6 text-2xl font-bold text-foreground">Wallet</h1>

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : isError || !wallet ? (
        <div className="flex flex-col items-center gap-4 rounded-3xl bg-surface p-10 text-center">
          <p className="text-sm text-muted-foreground">You don&apos;t have a wallet yet.</p>
          <Button onClick={() => initWalletMutation.mutate()} disabled={initWalletMutation.isPending}>
            {initWalletMutation.isPending ? "Setting up..." : "Set up my wallet"}
          </Button>
        </div>
      ) : (
        <>
          <BalanceCard balance={wallet.balance} onTopUp={() => setTopUpOpen(true)} onWithdraw={() => setWithdrawOpen(true)} />

          <div className="mt-6">
            <h2 className="mb-3 text-lg font-bold text-foreground">Transaction history</h2>
            <div className="space-y-0">
              {TRANSACTIONS_MOCK.map((transaction) => (
                <TransactionListItem key={transaction.txnId} transaction={transaction} />
              ))}
            </div>
          </div>

          <TopUpDialog
            open={topUpOpen}
            onOpenChange={setTopUpOpen}
            onSubmit={(amount) => topUpMutation.mutate(amount)}
            isSubmitting={topUpMutation.isPending}
          />
          <WithdrawDialog
            open={withdrawOpen}
            onOpenChange={setWithdrawOpen}
            onSubmit={(payload) => withdrawMutation.mutate(payload)}
            isSubmitting={withdrawMutation.isPending}
          />
        </>
      )}
    </div>
  );
}
