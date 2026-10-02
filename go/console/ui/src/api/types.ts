/**
 * TypeScript types for every JSON shape in console/docs/spec.md.
 * Field names mirror the Go backend exactly (snake_case).
 */

/** Response envelope carried by every backend reply. */
export interface Envelope<T> {
  /** True when the request succeeded and `data` is populated. */
  success: boolean;
  /** Payload on success. */
  data?: T;
  /** Human-readable error message on failure. */
  error?: string;
  /** Five-digit error code on failure (see {@link ErrorCode}). */
  code?: number;
}

/** Error codes returned by the backend. */
export const ErrorCode = {
  BadRequest: 40000,
  Unauthenticated: 40100,
  Forbidden: 40300,
  NotFound: 40400,
  Conflict: 40900,
  Validation: 42200,
  Internal: 50000,
  Upstream: 50200,
} as const;

/** Union of the known error code values. */
export type ErrorCodeValue = (typeof ErrorCode)[keyof typeof ErrorCode];

/** Extra payload attached to a 50200 upstream error (the SDK's APIError). */
export interface UpstreamErrorData {
  /** HTTP status CYBERBIZ returned. */
  status: number;
  /** Error messages CYBERBIZ returned. */
  messages: string[];
  /** CYBERBIZ request id, when present. */
  request_id: string;
}

/** A logged-in Console user. */
export interface User {
  id: number;
  username: string;
}

/** Reply of `GET /api/auth/me`. */
export interface MeResponse {
  user: User;
  /** True while no Shop exists; the UI shows the first-run Shop form. */
  setup_required: boolean;
}

/** Body of `POST /api/auth/login`. */
export interface LoginRequest {
  username: string;
  password: string;
}

/** Reply of `POST /api/auth/login`. */
export interface LoginResponse {
  user: User;
}

/** One CYBERBIZ merchant store with its App credentials. Secrets are never returned. */
export interface Shop {
  id: number;
  name: string;
  shop_domain: string;
  custom_domain: string;
  cyberbiz_shop_id: number;
  app_name: string;
  app_id: string;
  has_app_secret: boolean;
  /** `sha256:<8 hex> len=N` of the stored API Token. */
  token_fingerprint: string;
  /** `CONSOLE_PUBLIC_URL + "/webhooks/cyberbiz"`, empty when the public URL is unset. */
  webhook_url: string;
  created_at: string;
  updated_at: string;
}

/** Body of `POST /api/shops`. */
export interface CreateShopRequest {
  name: string;
  shop_domain: string;
  app_name: string;
  app_id: string;
  app_secret: string;
  api_token: string;
}

/** Body of `PUT /api/shops/:id`; blank secret/token keeps the stored one. */
export type UpdateShopRequest = Partial<CreateShopRequest>;

/** Reply of `POST /api/shops/:id/verify`: whatever CYBERBIZ returned for `GET /shop`. */
export interface VerifyShopResponse {
  shop_info: JsonValue;
}

/** Any JSON value. */
export type JsonValue =
  string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue };

/** A JSON object. */
export type JsonObject = Record<string, JsonValue>;

/** Where a {@link Param} lives in the request. */
export type ParamLocation = 'path' | 'query';

/** One path or query parameter of an {@link Operation}. */
export interface Param {
  name: string;
  in: ParamLocation;
  required: boolean;
  /** OpenAPI primitive type: string, integer, number, boolean, array. */
  type: string;
  description: string;
  /** Allowed values when the parameter is an enum. */
  enum: string[] | null;
  /** Example value from the OpenAPI document. */
  example: JsonValue | null;
}

/** Request body description of an {@link Operation}. */
export interface RequestBodySpec {
  content_type: string;
  schema: JsonValue;
  example: JsonValue | null;
}

/** One documented response of an {@link Operation}. */
export interface ResponseSpec {
  status: number;
  description: string;
  example: JsonValue | null;
}

/** HTTP methods the catalog can describe. */
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS';

/** One operation from the OpenAPI catalog. */
export interface Operation {
  id: string;
  method: HttpMethod;
  path: string;
  tag: string;
  summary: string;
  description: string;
  path_params: Param[];
  query_params: Param[];
  request_body: RequestBodySpec | null;
  responses: ResponseSpec[];
}

/** One tag grouping catalog operations. */
export interface Tag {
  name: string;
  description: string;
}

/** Reply of `GET /api/catalog`. */
export interface Catalog {
  operations: Operation[];
  tags: Tag[];
}

/** Body of `POST /api/shops/:id/requests`. */
export interface SendRequestBody {
  method: HttpMethod;
  path: string;
  query: Record<string, string[]>;
  body: JsonValue | null;
  headers: Record<string, string>;
}

/** Fields shared by the Outbound summary and detail shapes. */
export interface OutboundLogSummary {
  id: number;
  shop_id: number;
  method: HttpMethod;
  path: string;
  query: string;
  response_status: number;
  duration_ms: number;
  attempt: number;
  request_id: string;
  error: string | null;
  created_at: string;
}

/** An Outbound with request and response headers and bodies. */
export interface OutboundLog extends OutboundLogSummary {
  /** Redacted request headers, JSON object of header name to value(s). */
  request_headers: Record<string, string | string[]>;
  request_body: string;
  response_headers: Record<string, string | string[]>;
  /** Response body text; base64 when the response was binary. */
  response_body: string;
}

/** Verification status of an Inbound. */
export type InboundStatus = 'valid' | 'invalid_signature' | 'unknown_shop' | 'malformed';

/** Fields shared by the Inbound summary and detail shapes. */
export interface InboundLogSummary {
  id: number;
  shop_id: number | null;
  shop_domain: string;
  custom_domain: string;
  event: string;
  status: InboundStatus;
  signature_valid: boolean;
  domain_signature_valid: boolean | null;
  response_status: number;
  /** Id of an earlier Inbound carrying the same Signature, when this one is a duplicate. */
  duplicate_of: number | null;
  remote_addr: string;
  created_at: string;
}

/** An Inbound with its headers, body and signatures. */
export interface InboundLog extends InboundLogSummary {
  signature: string;
  domain_signature: string;
  headers: Record<string, string | string[]>;
  body: string;
}

/** Paginated list envelope used by the Outbound and Inbound list endpoints. */
export interface Paginated<T> {
  items: T[];
  total: number;
  page: number;
  per_page: number;
}

/** Query parameters accepted by `GET /api/outbound`. */
export interface OutboundFilter {
  shop_id?: number;
  method?: string;
  path?: string;
  status?: number;
  from?: string;
  to?: string;
  q?: string;
  page?: number;
  per_page?: number;
}

/** Query parameters accepted by `GET /api/inbound`. */
export interface InboundFilter {
  shop_id?: number;
  event?: string;
  status?: InboundStatus;
  from?: string;
  to?: string;
  q?: string;
  page?: number;
  per_page?: number;
}

/** Body of `POST /api/{outbound,inbound}/:id/golden`. */
export interface SaveGoldenRequest {
  /** Optional name suffix appended to the generated file name. */
  name?: string;
  /** Replace an existing Golden File instead of refusing with 40900. */
  overwrite?: boolean;
}

/** Reply of the golden endpoints. */
export interface SaveGoldenResponse {
  path: string;
}

/** Reply of `GET /api/stats`. */
export interface Stats {
  shops: number;
  outbound_24h: number;
  inbound_24h: number;
  inbound_invalid_24h: number;
}

/** Reply of `GET /api/health`. */
export interface Health {
  status: string;
  version: string;
}
