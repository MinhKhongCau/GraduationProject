export interface ApiErrorResponse {
  message: string;
  statusCode?: number;
  details?: unknown;
}

/** Matches profile-service's schemas.PaginatedResponse (Go services use this shape). */
export interface PaginatedResponse<T> {
  items: T[];
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
}

export interface MessageResponse {
  message: string;
}

/** profile-service/payment-service/booking-service wrap responses as { message, data }. */
export interface ServiceEnvelope<T> {
  message: string;
  data: T;
}
