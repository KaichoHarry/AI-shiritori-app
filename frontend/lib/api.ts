import type {
  Difficulty,
  MessageRecord,
  MessageSubmitResult,
  NextHint,
  Session,
  Settings,
  Tone,
  User,
  WordRecord,
  WordSubmitResult,
} from "./types";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  reason: string;
  status: number;

  constructor(status: number, reason: string, message: string) {
    super(message);
    this.status = status;
    this.reason = reason;
  }
}

type FetchOptions = {
  method?: string;
  body?: unknown;
  token?: string;
};

async function apiFetch<T>(path: string, options: FetchOptions = {}): Promise<T> {
  const headers: Record<string, string> = {};
  if (options.body !== undefined) {
    headers["Content-Type"] = "application/json";
  }
  if (options.token) {
    headers["Authorization"] = `Bearer ${options.token}`;
  }

  const res = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method ?? "GET",
    headers,
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
  });

  const isJson = res.headers.get("content-type")?.includes("application/json");
  const data = isJson ? await res.json().catch(() => ({})) : {};

  if (!res.ok) {
    throw new ApiError(
      res.status,
      (data as { reason?: string }).reason ?? "unknown_error",
      (data as { message?: string }).message ?? "エラーが発生しました",
    );
  }

  return data as T;
}

export type AuthTokens = {
  access_token: string;
  refresh_token: string;
};

export type LoginResponse = AuthTokens & { user: User };

export function register(email: string, password: string, displayName: string) {
  return apiFetch<User>("/api/auth/register", {
    method: "POST",
    body: { email, password, display_name: displayName },
  });
}

export function login(email: string, password: string) {
  return apiFetch<LoginResponse>("/api/auth/login", {
    method: "POST",
    body: { email, password },
  });
}

export function refreshTokens(refreshToken: string) {
  return apiFetch<AuthTokens>("/api/auth/refresh", {
    method: "POST",
    body: { refresh_token: refreshToken },
  });
}

export function fetchMe(token: string) {
  return apiFetch<User>("/api/auth/me", { token });
}

export function requestPasswordReset(email: string) {
  return apiFetch<{ message: string }>("/api/auth/password-reset/request", {
    method: "POST",
    body: { email },
  });
}

export function getSettings(token: string) {
  return apiFetch<Settings>("/api/settings", { token });
}

export function updateSettings(
  token: string,
  mode2Difficulty: Difficulty,
  mode3Tone: Tone,
) {
  return apiFetch<Settings>("/api/settings", {
    method: "PUT",
    token,
    body: { mode2_difficulty: mode2Difficulty, mode3_tone: mode3Tone },
  });
}

export function createSession(
  token: string,
  mode: string,
  difficulty?: string,
  tone?: string,
) {
  return apiFetch<Session>("/api/games", {
    method: "POST",
    token,
    body: { mode, difficulty: difficulty ?? "", tone: tone ?? "" },
  });
}

export function listSessions(token: string, limit = 20, offset = 0) {
  return apiFetch<{ sessions: Session[] }>(
    `/api/games?limit=${limit}&offset=${offset}`,
    { token },
  );
}

export function getSessionDetail(token: string, id: string) {
  return apiFetch<{
    session: Session;
    words?: WordRecord[];
    messages?: MessageRecord[];
    next_hint?: NextHint;
  }>(`/api/games/${id}`, { token });
}

export function endSession(token: string, id: string) {
  return apiFetch<Session>(`/api/games/${id}/end`, { method: "POST", token });
}

export function submitWord(token: string, id: string, word: string) {
  return apiFetch<WordSubmitResult>(`/api/games/${id}/words`, {
    method: "POST",
    token,
    body: { word },
  });
}

export function submitMessage(token: string, id: string, content: string) {
  return apiFetch<MessageSubmitResult>(`/api/games/${id}/messages`, {
    method: "POST",
    token,
    body: { content },
  });
}
