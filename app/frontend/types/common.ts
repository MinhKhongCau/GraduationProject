export interface ApiErrorResponse {
  message: string;
  statusCode?: number;
  details?: unknown;
}

/** App-wide paginated list shape; API modules map each service's page format into it. */
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

/**
 * `result` of a paginated profile-service response (BaseResponse envelope is
 * unwrapped by the http client, see unwrapEnvelope: "result").
 */
export interface PageResult<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  hasNext: boolean;
  hasPrevious: boolean;
}

/** payment-service/booking-service wrap responses as { message, data }. */
export interface ServiceEnvelope<T> {
  message: string;
  data: T;
}
