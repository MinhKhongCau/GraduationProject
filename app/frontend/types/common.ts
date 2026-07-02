export interface ApiErrorResponse {
  message: string;
  statusCode?: number;
  details?: unknown;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface MessageResponse {
  message: string;
}

/** profile-service/payment-service/booking-service wrap responses as { message, data }. */
export interface ServiceEnvelope<T> {
  message: string;
  data: T;
}
