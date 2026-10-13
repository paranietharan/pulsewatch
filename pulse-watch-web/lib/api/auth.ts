import { apiClient, setAccessToken, performTokenRefresh } from "./client";
import {
  AuthResponse,
  SignupResponse,
  SafeUser,
  MessageResponse,
} from "@/types/api";

export async function login(payload: {
  email: string;
  password: string;
}): Promise<AuthResponse> {
  const res = await apiClient<AuthResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  setAccessToken(res.access_token);
  return res;
}

export async function signup(payload: {
  first_name: string;
  last_name: string;
  email: string;
  password: string;
}): Promise<SignupResponse> {
  return apiClient<SignupResponse>("/api/v1/auth/signup", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function refresh(): Promise<AuthResponse | null> {
  return performTokenRefresh();
}

export async function logout(): Promise<MessageResponse> {
  try {
    return await apiClient<MessageResponse>("/api/v1/auth/logout", {
      method: "POST",
    });
  } finally {
    setAccessToken(null);
  }
}

export async function getMe(): Promise<{ user: SafeUser }> {
  return apiClient<{ user: SafeUser }>("/api/v1/auth/me", {
    method: "GET",
  });
}

export async function verifyEmail(token: string): Promise<MessageResponse> {
  return apiClient<MessageResponse>("/api/v1/auth/verify-email", {
    method: "POST",
    body: JSON.stringify({ token }),
  });
}

export async function resendVerification(email: string): Promise<MessageResponse> {
  return apiClient<MessageResponse>("/api/v1/auth/resend-verification", {
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export async function forgotPassword(email: string): Promise<MessageResponse> {
  return apiClient<MessageResponse>("/api/v1/auth/forgot-password", {
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export async function resetPassword(payload: {
  token: string;
  new_password: string;
}): Promise<MessageResponse> {
  return apiClient<MessageResponse>("/api/v1/auth/reset-password", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}
