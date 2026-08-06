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

export function authHeaders(): Record<string, string> {
  const token = localStorage.getItem('jwt_token');
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

export async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const resp = await fetch(url, {
    headers: authHeaders(),
    ...options,
  });
  if (!resp.ok) {
    throw new Error(`Request failed: ${resp.status}`);
  }
  return resp.json();
}

export async function requestVoid(url: string, options?: RequestInit): Promise<void> {
  const resp = await fetch(url, {
    headers: authHeaders(),
    ...options,
  });
  if (!resp.ok) {
    throw new Error(`Request failed: ${resp.status}`);
  }
}

export async function requestRaw(url: string, options?: RequestInit): Promise<Response> {
  const resp = await fetch(url, {
    headers: authHeaders(),
    ...options,
  });
  if (!resp.ok) {
    throw new Error(`Request failed: ${resp.status}`);
  }
  return resp;
}
