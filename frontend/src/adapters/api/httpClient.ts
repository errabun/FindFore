const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '';

export const endpoints = {
  players: `${API_BASE}/api/v1/players`,
  courses: `${API_BASE}/api/v1/courses`,
  playerEvent: `${API_BASE}/api/v1/player-event`,
  joinEvent: `${API_BASE}/api/v1/player-event/join`,
  singleEvent: `${API_BASE}/api/v1/event`,
  friendship: `${API_BASE}/api/v1/friendship`,
  friendships: `${API_BASE}/api/v1/friendships`,
  posts: `${API_BASE}/api/v1/posts`,
  sessions: `${API_BASE}/api/v1/sessions`,
};

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

type ErrorEnvelope = {
  errors?: Array<{ code?: string; message?: string }>;
};

let onUnauthorized: (() => void) | null = null;

/** Register a callback invoked when any authenticated request returns 401. */
export function setUnauthorizedHandler(handler: (() => void) | null) {
  onUnauthorized = handler;
}

export function authHeaders(): Record<string, string> {
  const token = localStorage.getItem('jwt_token');
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

async function parseApiError(resp: Response): Promise<ApiError> {
  let code = 'request_failed';
  let message = `Request failed: ${resp.status}`;
  try {
    const body = (await resp.json()) as ErrorEnvelope;
    const first = body.errors?.[0];
    if (first?.code) code = first.code;
    if (first?.message) message = first.message;
  } catch {
    // keep defaults
  }
  return new ApiError(resp.status, code, message);
}

async function handleResponse<T>(resp: Response, { allowUnauthorized = false } = {}): Promise<T> {
  if (resp.status === 401 && !allowUnauthorized) {
    onUnauthorized?.();
  }
  if (!resp.ok) {
    throw await parseApiError(resp);
  }
  if (resp.status === 204) {
    return undefined as T;
  }
  const text = await resp.text();
  if (!text) {
    return undefined as T;
  }
  return JSON.parse(text) as T;
}

export async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const resp = await fetch(url, {
    headers: authHeaders(),
    ...options,
  });
  return handleResponse<T>(resp);
}

export async function requestVoid(url: string, options?: RequestInit): Promise<void> {
  const resp = await fetch(url, {
    headers: authHeaders(),
    ...options,
  });
  await handleResponse<void>(resp);
}

export async function requestRaw(url: string, options?: RequestInit): Promise<Response> {
  const resp = await fetch(url, {
    headers: authHeaders(),
    ...options,
  });
  if (!resp.ok) {
    throw await parseApiError(resp);
  }
  return resp;
}

/** Public auth endpoints should not trigger global logout on 401. */
export async function requestPublic<T>(url: string, options?: RequestInit): Promise<T> {
  const resp = await fetch(url, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  return handleResponse<T>(resp, { allowUnauthorized: true });
}
