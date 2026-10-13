"use client";

import React, { createContext, useContext, useEffect, useState, useCallback } from "react";
import { SafeUser } from "@/types/api";
import * as authApi from "@/lib/api/auth";
import { setAccessToken, performTokenRefresh, onSessionExpired } from "@/lib/api/client";

interface AuthContextType {
  user: SafeUser | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<boolean>;
  setUser: React.Dispatch<React.SetStateAction<SafeUser | null>>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<SafeUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Clear state when session expires
  useEffect(() => {
    const unsubscribe = onSessionExpired(() => {
      setToken(null);
      setUser(null);
    });
    return unsubscribe;
  }, []);

  const refreshSession = useCallback(async (): Promise<boolean> => {
    try {
      const res = await performTokenRefresh();
      if (res && res.access_token && res.user) {
        setToken(res.access_token);
        setUser(res.user);
        return true;
      }
      setToken(null);
      setUser(null);
      return false;
    } catch {
      setToken(null);
      setUser(null);
      return false;
    }
  }, []);

  useEffect(() => {
    let isMounted = true;
    (async () => {
      try {
        await refreshSession();
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    })();
    return () => {
      isMounted = false;
    };
  }, [refreshSession]);

  const login = async (email: string, password: string) => {
    const res = await authApi.login({ email, password });
    setToken(res.access_token);
    setUser(res.user);
    setAccessToken(res.access_token);
  };

  const logout = async () => {
    try {
      await authApi.logout();
    } finally {
      setToken(null);
      setUser(null);
      setAccessToken(null);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isLoading,
        isAuthenticated: !!user,
        login,
        logout,
        refreshSession,
        setUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
