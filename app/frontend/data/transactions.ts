import type { Transaction } from "@/types";

/** payment-service has no GET transaction-history endpoint yet — mock only. */
export const TRANSACTIONS_MOCK: Transaction[] = [
  {
    txnId: "txn-1",
    walletId: "wallet-mock",
    transactionType: "TOP_UP",
    amount: 2000000,
    balanceBefore: 13450000,
    balanceAfter: 15450000,
    status: "SUCCESS",
    createdAt: new Date(2026, 5, 20, 14, 20).toISOString(),
  },
  {
    txnId: "txn-2",
    walletId: "wallet-mock",
    relatedOrderId: "appt-seed-1",
    transactionType: "PAYMENT",
    amount: -500000,
    balanceBefore: 13950000,
    balanceAfter: 13450000,
    status: "SUCCESS",
    createdAt: new Date(2026, 5, 20, 9, 15).toISOString(),
  },
  {
    txnId: "txn-3",
    walletId: "wallet-mock",
    transactionType: "WITHDRAW",
    amount: -1000000,
    balanceBefore: 14950000,
    balanceAfter: 13950000,
    status: "PENDING",
    createdAt: new Date(2026, 5, 19, 18, 30).toISOString(),
  },
];
