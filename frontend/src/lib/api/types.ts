// Shared response envelope types mirrored from
// backend/internal/presentation/http/response/common.go. Every group file
// under src/lib/api/ reuses these instead of redeclaring them.

export interface PageResponse {
  page: number;
  limit: number;
  total_items: number;
}

export interface PageDataResponse<T> {
  data: T[];
  page: PageResponse;
}

export interface CountDataResponse<T> {
  data: T[];
  total_items: number;
}

export interface IdResponse {
  id: string;
}

export interface CountResponse {
  count: number;
}

export interface ErrorResponse {
  error: string;
  message: string;
}

// Mirrors response/common.go's AuditResponse, embedded (via Go struct
// embedding) into most resource response types. `updated_at`/`deleted_at`/
// `created_by`/`updated_by`/`deleted_by` are `omitempty` on the wire, so
// they're optional here too.
export interface AuditFields {
  created_at: string;
  updated_at?: string;
  deleted_at?: string;
  created_by?: string;
  updated_by?: string;
  deleted_by?: string;
}

// Common list query params shared by every paginated (PageDataResponse)
// list endpoint - built via buildQuery() in client.ts.
export interface PageQuery {
  page?: number;
  limit?: number;
  search?: string;
}
