export type WalletOwnerType = "PATIENT" | "EXPERT" | "SYSTEM";

export interface Wallet {
  walletId: string;
  ownerId: string;
  userType: WalletOwnerType;
  balance: number;
  updatedAt: string;
}

export type TransactionType =
  | "TOP_UP"
  | "PAYMENT"
  | "EARNING"
  | "COMMISSION"
  | "WITHDRAW"
  | "REFUND_WITHDRAW";

export type TransactionStatus = "PENDING" | "SUCCESS" | "FAILED";

export interface Transaction {
  txnId: string;
  walletId: string;
  relatedOrderId?: string;
  transactionType: TransactionType;
  amount: number;
  balanceBefore: number;
  balanceAfter: number;
  status: TransactionStatus;
  createdAt: string;
}

export type WithdrawalStatus = "PENDING" | "APPROVED" | "REJECTED";

export interface WithdrawalRequestRecord {
  requestId: string;
  walletId: string;
  amount: number;
  bankInfo: string;
  status: WithdrawalStatus;
  adminNote?: string;
  createdAt: string;
}

export interface InitWalletRequest {
  ownerId: string;
  userType: WalletOwnerType;
}

export interface TopUpRequest {
  amount: number;
}

export interface PaymentRequest {
  payerId: string;
  payeeId: string;
  amount: number;
  relatedOrderId: string;
}

export interface CreateWithdrawalRequest {
  amount: number;
  bankInfo: string;
}

export interface ProcessWithdrawalRequest {
  action: "APPROVE" | "REJECT";
  adminNote?: string;
}

export type PaymentGateway = "VNPAY" | "MOMO" | "MOCK";
export type OrderStatus = "PENDING" | "SUCCESS" | "FAILED" | "EXPIRED";

export interface CreateOrderRequest {
  expertId: string;
  amount: number;
  gateway: PaymentGateway;
  /** Ties the order to a booking-service appointment; server recomputes amount from the real slot price when set. */
  appointmentId?: string;
}

export type PaymentReturnBookingStatus =
  | "PENDING_PAYMENT"
  | "CONFIRMED"
  | "CANCELLED"
  | "COMPLETED"
  | "UNKNOWN";

/** Result of GET /payments/vnpay-return (payment-service). */
export interface PaymentReturnResult {
  orderId: string;
  appointmentId?: string;
  paymentStatus: OrderStatus;
  /** Present only when the order belongs to an appointment. */
  bookingStatus?: PaymentReturnBookingStatus;
  amountVnd: number;
  /** Raw vnp_ResponseCode ("00" = success, "24" = cancelled by user, ...). */
  responseCode: string;
  gatewayTxnRef?: string;
  paidAt?: number;
}

export interface CreateOrderResponse {
  orderId: string;
  grossAmount: number;
  netAmount: number;
  commissionAmount: number;
  paymentUrl: string;
  status: OrderStatus;
}
