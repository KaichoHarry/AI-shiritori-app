"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

import {
  ApiError,
  fetchMe,
  login as apiLogin,
  refreshTokens,
  register as apiRegister,
} from "@/lib/api";
import type { User } from "@/lib/types";

const STORAGE_KEY = "shiritori.auth";

type StoredAuth = {
  accessToken: string;
  refreshToken: string;
};

type AuthContextValue = {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (
    email: string,
    password: string,
    displayName: string,
  ) => Promise<void>;
  logout: () => void;
  /**
   * 認証付きAPI呼び出しのラッパー。アクセストークンが期限切れ(401)の場合、
   * リフレッシュトークンで1回だけ再試行する。
   */
  authFetch: <T>(fn: (token: string) => Promise<T>) => Promise<T>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function loadStoredAuth(): StoredAuth | null {
  if (typeof window === "undefined") return null;
  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as StoredAuth;
  } catch {
    return null;
  }
}

function saveStoredAuth(auth: StoredAuth | null) {
  if (typeof window === "undefined") return;
  if (auth) {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(auth));
  } else {
    window.localStorage.removeItem(STORAGE_KEY);
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [initialAuth] = useState(() => loadStoredAuth());
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(
    initialAuth?.accessToken ?? null,
  );
  const [refreshToken, setRefreshToken] = useState<string | null>(
    initialAuth?.refreshToken ?? null,
  );
  const [loading, setLoading] = useState(!!initialAuth);

  useEffect(() => {
    if (!initialAuth) return;
    fetchMe(initialAuth.accessToken)
      .then((u) => setUser(u))
      .catch(() => {
        saveStoredAuth(null);
        setAccessToken(null);
        setRefreshToken(null);
        setUser(null);
      })
      .finally(() => setLoading(false));
  }, [initialAuth]);

  const applyTokens = useCallback((access: string, refresh: string) => {
    setAccessToken(access);
    setRefreshToken(refresh);
    saveStoredAuth({ accessToken: access, refreshToken: refresh });
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await apiLogin(email, password);
      applyTokens(res.access_token, res.refresh_token);
      setUser(res.user);
    },
    [applyTokens],
  );

  const register = useCallback(
    async (email: string, password: string, displayName: string) => {
      await apiRegister(email, password, displayName);
      await login(email, password);
    },
    [login],
  );

  const logout = useCallback(() => {
    setUser(null);
    setAccessToken(null);
    setRefreshToken(null);
    saveStoredAuth(null);
  }, []);

  const authFetch = useCallback(
    async <T,>(fn: (token: string) => Promise<T>): Promise<T> => {
      if (!accessToken) {
        throw new ApiError(401, "unauthorized", "ログインしてください");
      }
      try {
        return await fn(accessToken);
      } catch (err) {
        if (err instanceof ApiError && err.status === 401 && refreshToken) {
          const tokens = await refreshTokens(refreshToken);
          applyTokens(tokens.access_token, tokens.refresh_token);
          return await fn(tokens.access_token);
        }
        throw err;
      }
    },
    [accessToken, refreshToken, applyTokens],
  );

  return (
    <AuthContext.Provider
      value={{ user, loading, login, register, logout, authFetch }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
