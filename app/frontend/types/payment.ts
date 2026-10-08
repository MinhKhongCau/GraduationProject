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

// ---------------------------------------------------------------------------
// Transaction management (payment-service /orders, /expert/orders, /admin/orders)
// ---------------------------------------------------------------------------

/** APPOINTMENT = paid booking, TOP_UP = wallet top-up (no appointment, no expert). */
export type PaymentOrderType = "APPOINTMENT" | "TOP_UP";

export type FulfillmentStatus =
  | "PENDING"
  | "BOOKING_CONFIRMED"
  | "BOOKING_FAILED"
  | "MANUAL_REVIEW"
  | "REFUND_REQUIRED";

export type GatewayCaptureStatus = "PENDING" | "CAPTURED" | "FAILED" | "CAPTURED_DUPLICATE";

/** payment-service's zero-based page envelope (readquery.Page). */
export interface PaymentPage<T> {
  items: T[];
  page: number;
  size: number;
  totalItems: number;
  totalPages: number;
  hasNext: boolean;
  hasPrevious: boolean;
}

/** Shared list filters; `from`/`to` are YYYY-MM-DD (server defaults to a 30-day window). */
export interface PaymentOrderFilters {
  status?: OrderStatus;
  type?: PaymentOrderType;
  fulfillmentStatus?: FulfillmentStatus;
  from?: string;
  to?: string;
  /** Zero-based. */
  page?: number;
  size?: number;
}

/** Admin can also narrow to one managed expert or one patient. */
export interface AdminPaymentOrderFilters extends PaymentOrderFilters {
  expertId?: string;
  payerId?: string;
}

export interface PatientPaymentOrder {
  id: string;
  appointmentId?: string;
  payerId: string;
  expertId: string;
  type: PaymentOrderType;
  amountVnd: number;
  status: OrderStatus;
  fulfillmentStatus: FulfillmentStatus;
  gatewayCaptureStatus: GatewayCaptureStatus;
  gateway: PaymentGateway;
  gatewayTransactionReference: string;
  expiresAt: number;
  createdAt: number;
  paidAt?: number;
}

export interface ExpertPaymentOrder {
  id: string;
  appointmentId?: string;
  payerId: string;
  type: PaymentOrderType;
  grossAmount: number;
  commissionRate: number;
  commissionAmount: number;
  netAmount: number;
  status: OrderStatus;
  fulfillmentStatus: FulfillmentStatus;
  released: boolean;
  createdAt: number;
  paidAt?: number;
}

export interface AdminPaymentOrder {
  id: string;
  appointmentId?: string;
  payerId: string;
  expertId: string;
  type: PaymentOrderType;
  grossAmount: number;
  commissionRate: number;
  commissionAmount: number;
  netAmount: number;
  gateway: PaymentGateway;
  gatewayTxnRef: string;
  gatewayResponseCode?: string;
  gatewayTransactionStatus?: string;
  gatewayPaymentDate?: string;
  status: OrderStatus;
  gatewayCaptureStatus: GatewayCaptureStatus;
  fulfillmentStatus: FulfillmentStatus;
  released: boolean;
  createdAt: number;
  expiresAt: number;
  paidAt?: number;
}

/** Money totals only count SUCCESS orders. `from`/`to` are the applied window in epoch ms. */
export interface PatientOrderSummary {
  totalOrders: number;
  successOrders: number;
  totalPaid: number;
  from: number;
  to: number;
}

export interface ExpertOrderSummary {
  totalOrders: number;
  successOrders: number;
  grossTotal: number;
  commissionTotal: number;
  netTotal: number;
  from: number;
  to: number;
}

export interface AdminOrderSummary extends ExpertOrderSummary {
  pendingOrders: number;
  failedOrders: number;
  expiredOrders: number;
  managedExperts: number;
}

export type CompensationStatus = "MANUAL_REVIEW" | "REFUND_REQUIRED" | "RESOLVED";
export type AdminReviewAction = Exclude<CompensationStatus, "RESOLVED">;

export interface CompensationCase {
  id: string;
  paymentOrderId: string;
  expertId: string;
  payerId: string;
  appointmentId: string;
  type: string;
  status: CompensationStatus;
  paymentStatus: OrderStatus;
  moneyPaid: boolean;
  fulfillmentStatus: FulfillmentStatus;
  reasonCode: string;
  safeReason: string;
  amountVnd: number;
  createdAt: number;
  updatedAt: number;
  resolvedAt?: number;
  resolutionNote?: string;
}

export interface CompensationCaseFilters {
  status?: CompensationStatus;
  expertId?: string;
  from?: string;
  to?: string;
  page?: number;
  size?: number;
}

/** Compensation list uses its own page shape (total + totalItems). */
export interface CompensationCasePage extends PaymentPage<CompensationCase> {
  total: number;
}

export type WalletTransactionType =
  | "PAYMENT_RECEIVED"
  | "COMMISSION_DEDUCTED"
  | "REFUND"
  | "WITHDRAWAL_LOCKED"
  | "WITHDRAWAL_COMPLETED"
  | "WITHDRAWAL_REJECTED"
  | "ADJUSTMENT";

export interface ManagedWalletTransaction {
  id: string;
  walletId: string;
  expertId: string;
  type: WalletTransactionType;
  /** Signed: > 0 credit, < 0 debit. */
  amount: number;
  balanceAfter: number;
  referenceType: string;
  referenceId: string;
  createdAt: number;
}

export interface ManagedWalletTransactionFilters {
  expertId?: string;
  type?: WalletTransactionType;
  direction?: "CREDIT" | "DEBIT";
  from?: string;
  to?: string;
  page?: number;
  size?: number;
}
