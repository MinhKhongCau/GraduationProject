import { paymentClient } from "./http/instances";
import { PAYMENT_ENDPOINTS } from "@/constants/api";
import type {
  Wallet,
  InitWalletRequest,
  TopUpRequest,
  PaymentRequest,
  CreateWithdrawalRequest,
  ProcessWithdrawalRequest,
  WithdrawalRequestRecord,
  ServiceEnvelope,
  CreateOrderRequest,
  CreateOrderResponse,
  PaymentReturnResult,
  PaymentPage,
  PaymentOrderFilters,
  AdminPaymentOrderFilters,
  PatientPaymentOrder,
  PatientOrderSummary,
  ExpertPaymentOrder,
  ExpertOrderSummary,
  AdminPaymentOrder,
  AdminOrderSummary,
  AdminReviewAction,
  CompensationCase,
  CompensationCaseFilters,
  CompensationCasePage,
  ManagedWalletTransaction,
  ManagedWalletTransactionFilters,
} from "@/types";

// payment-service is another Go/Gin service like profile-service; assuming
// the same { message, data } envelope convention (unconfirmed in source,
// adjust if it turns out to return the entity directly).

export async function initWallet(payload: InitWalletRequest): Promise<Wallet> {
  const response = await paymentClient.get<ServiceEnvelope<any>>(
    PAYMENT_ENDPOINTS.INIT_WALLET
  );
  const raw = response.data?.data;
  return {
    walletId: raw?.id ?? "",
    ownerId: raw?.userId ?? payload.ownerId,
    userType: payload.userType,
    balance: raw?.availableBalance ?? 0,
    updatedAt: raw?.updatedAt ? new Date(raw.updatedAt).toISOString() : new Date().toISOString(),
  };
}

export async function getWallet(ownerId: string): Promise<Wallet> {
  const response = await paymentClient.get<ServiceEnvelope<any>>(
    PAYMENT_ENDPOINTS.WALLET
  );
  const raw = response.data?.data;
  return {
    walletId: raw?.id ?? "",
    ownerId: raw?.userId ?? ownerId,
    userType: "PATIENT",
    balance: raw?.availableBalance ?? 0,
    updatedAt: raw?.updatedAt ? new Date(raw.updatedAt).toISOString() : new Date().toISOString(),
  };
}

export async function topUpWallet(ownerId: string, payload: TopUpRequest): Promise<Wallet> {
  const response = await paymentClient.post<ServiceEnvelope<any>>(
    PAYMENT_ENDPOINTS.TOP_UP,
    payload
  );
  const raw = response.data?.data;
  return {
    walletId: raw?.id ?? "",
    ownerId: raw?.userId ?? ownerId,
    userType: "PATIENT",
    balance: raw?.availableBalance ?? 0,
    updatedAt: raw?.updatedAt ? new Date(raw.updatedAt).toISOString() : new Date().toISOString(),
  };
}

export async function pay(payload: PaymentRequest): Promise<void> {
  await paymentClient.post(PAYMENT_ENDPOINTS.PAY, payload);
}

export async function requestWithdrawal(
  ownerId: string,
  payload: CreateWithdrawalRequest
): Promise<WithdrawalRequestRecord> {
  const response = await paymentClient.post<ServiceEnvelope<WithdrawalRequestRecord>>(
    PAYMENT_ENDPOINTS.WITHDRAW,
    payload
  );
  return response.data.data;
}

export async function processWithdrawal(
  requestId: string,
  payload: ProcessWithdrawalRequest
): Promise<WithdrawalRequestRecord> {
  const response = await paymentClient.post<ServiceEnvelope<WithdrawalRequestRecord>>(
    PAYMENT_ENDPOINTS.PROCESS_WITHDRAWAL(requestId),
    payload
  );
  return response.data.data;
}

/** Creates a VNPay order for a booking; redirect the patient to the returned paymentUrl. */
export async function createOrder(payload: CreateOrderRequest): Promise<CreateOrderResponse> {
  const response = await paymentClient.post<ServiceEnvelope<CreateOrderResponse>>(
    PAYMENT_ENDPOINTS.ORDERS,
    payload
  );
  return response.data.data;
}

/**
 * Forwards the query string VNPay appended to vnp_ReturnUrl so payment-service can verify the
 * signature, settle the order and report payment + booking status. The query is passed verbatim
 * in the URL (not via `params`) because the client's snake_case transform would alter the signed keys.
 */
export async function verifyVNPayReturn(rawQuery: string): Promise<PaymentReturnResult> {
  const query = rawQuery.startsWith("?") ? rawQuery.slice(1) : rawQuery;
  const response = await paymentClient.get<ServiceEnvelope<PaymentReturnResult>>(
    `${PAYMENT_ENDPOINTS.VNPAY_RETURN}?${query}`
  );
  return response.data.data;
}

// ---------------------------------------------------------------------------
// Transaction management. Lists are zero-based pages; money totals only count
// SUCCESS orders. Admin endpoints only return data of experts the admin
// approved (= manages) and answer 403 for anyone else's expert.
// ---------------------------------------------------------------------------

async function getData<T>(url: string, params?: object): Promise<T> {
  const response = await paymentClient.get<ServiceEnvelope<T>>(url, { params });
  return response.data.data;
}

/** [PATIENT] Own payment orders. */
export function listMyPaymentOrders(filters: PaymentOrderFilters = {}) {
  return getData<PaymentPage<PatientPaymentOrder>>(PAYMENT_ENDPOINTS.ORDERS, filters);
}

/** [PATIENT] Total paid within the filter window. */
export function getMyPaymentSummary(filters: PaymentOrderFilters = {}) {
  return getData<PatientOrderSummary>(PAYMENT_ENDPOINTS.ORDERS_SUMMARY, filters);
}

/** [EXPERT] Orders paid to the logged-in expert. */
export function listExpertPaymentOrders(filters: PaymentOrderFilters = {}) {
  return getData<PaymentPage<ExpertPaymentOrder>>(PAYMENT_ENDPOINTS.EXPERT_ORDERS, filters);
}

/** [EXPERT] Gross revenue, platform commission and net amount. */
export function getExpertPaymentSummary(filters: PaymentOrderFilters = {}) {
  return getData<ExpertOrderSummary>(PAYMENT_ENDPOINTS.EXPERT_ORDERS_SUMMARY, filters);
}

/** [ADMIN] Orders of managed experts. */
export function listAdminPaymentOrders(filters: AdminPaymentOrderFilters = {}) {
  return getData<PaymentPage<AdminPaymentOrder>>(PAYMENT_ENDPOINTS.ADMIN_ORDERS, filters);
}

export function getAdminPaymentSummary(filters: AdminPaymentOrderFilters = {}) {
  return getData<AdminOrderSummary>(PAYMENT_ENDPOINTS.ADMIN_ORDERS_SUMMARY, filters);
}

export function getAdminPaymentOrder(orderId: string) {
  return getData<AdminPaymentOrder>(PAYMENT_ENDPOINTS.ADMIN_ORDER(orderId));
}

/** [ADMIN] Flags a paid appointment order for manual review or refund; opens a compensation case. */
export async function reviewPaymentOrder(
  orderId: string,
  payload: { action: AdminReviewAction; note?: string }
): Promise<CompensationCase> {
  const response = await paymentClient.post<ServiceEnvelope<CompensationCase>>(
    PAYMENT_ENDPOINTS.ADMIN_ORDER_REVIEW(orderId),
    payload
  );
  return response.data.data;
}

/** [ADMIN] Compensation cases of managed experts. */
export function listCompensationCases(filters: CompensationCaseFilters = {}) {
  return getData<CompensationCasePage>(PAYMENT_ENDPOINTS.COMPENSATION_CASES, filters);
}

export async function resolveCompensationCase(caseId: string, note?: string): Promise<CompensationCase> {
  const response = await paymentClient.post<ServiceEnvelope<CompensationCase>>(
    PAYMENT_ENDPOINTS.COMPENSATION_CASE_RESOLVE(caseId),
    { note }
  );
  return response.data.data;
}

/** [ADMIN] Wallet ledger of managed experts. */
export function listManagedWalletTransactions(filters: ManagedWalletTransactionFilters = {}) {
  return getData<PaymentPage<ManagedWalletTransaction>>(PAYMENT_ENDPOINTS.ADMIN_WALLET_TRANSACTIONS, filters);
}
