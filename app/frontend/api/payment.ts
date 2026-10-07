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
