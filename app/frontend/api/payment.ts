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
} from "@/types";

// payment-service is another Go/Gin service like profile-service; assuming
// the same { message, data } envelope convention (unconfirmed in source,
// adjust if it turns out to return the entity directly).

export async function initWallet(payload: InitWalletRequest): Promise<Wallet> {
  const response = await paymentClient.get<ServiceEnvelope<Wallet>>(
    PAYMENT_ENDPOINTS.INIT_WALLET
  );
  return response.data.data;
}

export async function getWallet(ownerId: string): Promise<Wallet> {
  const response = await paymentClient.get<ServiceEnvelope<Wallet>>(
    PAYMENT_ENDPOINTS.WALLET
  );
  return response.data.data;
}

export async function topUpWallet(ownerId: string, payload: TopUpRequest): Promise<Wallet> {
  const response = await paymentClient.post<ServiceEnvelope<Wallet>>(
    PAYMENT_ENDPOINTS.TOP_UP,
    payload
  );
  return response.data.data;
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
