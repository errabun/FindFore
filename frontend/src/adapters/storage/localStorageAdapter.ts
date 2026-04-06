const JWT_KEY = 'jwt_token';
const ACTIVITY_KEY = 'last_activity';
const SESSION_TIMEOUT_MS = 30 * 60 * 1000; // 30 minutes

export function getToken(): string | null {
  return localStorage.getItem(JWT_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(JWT_KEY, token);
}

export function removeToken(): void {
  localStorage.removeItem(JWT_KEY);
}

export function getPlayerIdFromToken(): number {
  try {
    const token = getToken();
    if (!token) return 0;
    const payload = JSON.parse(atob(token.split('.')[1]));
    if (payload.exp * 1000 < Date.now()) return 0;
    return payload.player_id || 0;
  } catch {
    return 0;
  }
}

export function isSessionActive(): boolean {
  const last = localStorage.getItem(ACTIVITY_KEY);
  if (!last) return false;
  return Date.now() - Number(last) < SESSION_TIMEOUT_MS;
}

export function touchActivity(): void {
  localStorage.setItem(ACTIVITY_KEY, String(Date.now()));
}

export function clearSession(): void {
  removeToken();
  localStorage.removeItem(ACTIVITY_KEY);
}

// Color scheme
const COLOR_SCHEME_KEY = 'ff_color_scheme';

export type ColorScheme = 'light' | 'dark' | 'auto';

export function getColorScheme(): ColorScheme {
  return (localStorage.getItem(COLOR_SCHEME_KEY) as ColorScheme) || 'light';
}

export function setColorScheme(scheme: ColorScheme): void {
  localStorage.setItem(COLOR_SCHEME_KEY, scheme);
}
