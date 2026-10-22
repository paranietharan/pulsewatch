import { ApiError, AuthResponse } from "@/types/api";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/$/, "") ||
  (typeof window !== "undefined"
    ? ""
    : process.env.INTERNAL_API_URL || "http://localhost:8080");

let memoryAccessToken: string | null = null;
let refreshPromise: Promise<AuthResponse | null> | null = null;

type SessionExpiredListener = () => void;
const sessionExpiredListeners: Set<SessionExpiredListener> = new Set();

export function getAccessToken(): string | null {
  return memoryAccessToken;
}

export function setAccessToken(token: string | null) {
  memoryAccessToken = token;
}

export function onSessionExpired(listener: SessionExpiredListener): () => void {
  sessionExpiredListeners.add(listener);
  return () => {
    sessionExpiredListeners.delete(listener);
  };
}

function notifySessionExpired() {
  setAccessToken(null);
  sessionExpiredListeners.forEach((listener) => {
    try {
      listener();
    } catch {
      // Ignore listener error
    }
  });
}

export class FetchError extends Error {
  status: number;
  data?: ApiError;

  constructor(status: number, message: string, data?: ApiError) {
    super(message);
    this.name = "FetchError";
    this.status = status;
    this.data = data;
  }
}

/**
 * Single-flight token refresher. All callers (client interceptor or AuthContext)
 * share this single execution so rotating refresh tokens are never double-consumed.
 */
export async function performTokenRefresh(): Promise<AuthResponse | null> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
      });

      if (!res.ok) {
        notifySessionExpired();
        return null;
      }

      const data: AuthResponse = await res.json();
      if (data && data.access_token) {
        setAccessToken(data.access_token);
        return data;
      }

      notifySessionExpired();
      return null;
    } catch {
      notifySessionExpired();
      return null;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

export async function apiClient<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = endpoint.startsWith("http") ? endpoint : `${API_BASE_URL}${endpoint}`;
  const headers = new Headers(options.headers || {});

  if (!headers.has("Content-Type") && !(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  if (memoryAccessToken && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${memoryAccessToken}`);
  }

  const response = await fetch(url, {
    ...options,
    credentials: "include",
    headers,
  });

  // Automatically refresh and retry once on 401 for non-auth endpoints
  if (response.status === 401 && !endpoint.includes("/auth/")) {
    const refreshed = await performTokenRefresh();
    if (refreshed && refreshed.access_token) {
      headers.set("Authorization", `Bearer ${refreshed.access_token}`);
      const retryResponse = await fetch(url, {
        ...options,
        credentials: "include",
        headers,
      });

      if (!retryResponse.ok) {
        const errorData = await retryResponse.json().catch(() => null);
        throw new FetchError(
          retryResponse.status,
          errorData?.message ||
            errorData?.error ||
            `Request failed with status ${retryResponse.status}`,
          errorData
        );
      }

      if (retryResponse.status === 204) {
        return {} as T;
      }

      return retryResponse.json();
    } else {
      // Refresh failed (token expired or revoked)
      notifySessionExpired();
    }
  }

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    const message =
      errorData?.message ||
      errorData?.error ||
      `Request failed with status ${response.status}`;
    throw new FetchError(response.status, message, errorData);
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}
