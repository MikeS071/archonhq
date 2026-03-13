export interface ApiErrorPayload {
  error?: {
    code?: string;
    message?: string;
    details?: Record<string, unknown>;
    correlation_id?: string;
  };
}

export class ApiError extends Error {
  status: number;
  code?: string;
  correlationId?: string;
  details?: Record<string, unknown>;

  constructor(status: number, message: string, payload?: ApiErrorPayload) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = payload?.error?.code;
    this.correlationId = payload?.error?.correlation_id;
    this.details = payload?.error?.details;
  }
}

function baseUrl(): string {
  const configured = (import.meta.env.PUBLIC_ARCHON_API_BASE_URL as string | undefined)?.trim();
  if (!configured) {
    return "";
  }
  return configured.endsWith("/") ? configured.slice(0, -1) : configured;
}

function authToken(): string {
  const configured = (import.meta.env.PUBLIC_ARCHON_API_TOKEN as string | undefined)?.trim();
  if (configured && configured.length > 0) {
    return configured;
  }

  if (typeof window !== "undefined") {
    return window.localStorage.getItem("archonhq.apiToken")?.trim() ?? "";
  }

  return "";
}

export async function apiGet<T>(path: string): Promise<T> {
  const url = `${baseUrl()}${path}`;
  const token = authToken();
  const headers = new Headers({
    accept: "application/json"
  });

  if (token) {
    headers.set("authorization", `Bearer ${token}`);
  }

  const response = await fetch(url, {
    method: "GET",
    headers
  });

  const text = await response.text();
  const payload = text ? (JSON.parse(text) as ApiErrorPayload & T) : ({} as ApiErrorPayload & T);

  if (!response.ok) {
    const message = payload.error?.message ?? `Request failed with status ${response.status}`;
    throw new ApiError(response.status, message, payload);
  }

  return payload as T;
}
