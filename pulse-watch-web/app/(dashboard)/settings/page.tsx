"use client";

import React, { useState } from "react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { Button } from "@/components/ui/Button";
import { useAuth } from "@/context/AuthContext";
import { useTheme } from "@/components/theme/ThemeProvider";
import {
  User,
  Palette,
  Server,
  LogOut,
  CheckCircle2,
  Moon,
  Sun,
  Monitor as MonitorIcon,
  Check,
} from "lucide-react";

export default function SettingsPage() {
  const { user, logout } = useAuth();
  const { theme, setTheme } = useTheme();
  const [loggingOut, setLoggingOut] = useState(false);

  const handleLogout = async () => {
    setLoggingOut(true);
    try {
      await logout();
    } finally {
      setLoggingOut(false);
    }
  };

  const apiBaseUrl =
    process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

  return (
    <DashboardLayout
      title="Settings"
      description="Manage account details, active session preferences, and client configuration"
    >
      <div className="flex flex-col gap-5 max-w-4xl">
        {/* Profile Card */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-5 shadow-2xs">
          <div className="flex items-center gap-2 pb-3.5 border-b border-neutral-100 dark:border-neutral-800/80 mb-4">
            <User className="w-4 h-4 text-emerald-500" />
            <h2 className="text-xs font-bold uppercase tracking-wider text-neutral-500">
              Account Profile
            </h2>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
            <div>
              <span className="text-neutral-500 block text-[11px] mb-1">
                Full Name
              </span>
              <p className="font-semibold text-neutral-900 dark:text-white">
                {user ? `${user.first_name} ${user.last_name}` : "—"}
              </p>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px] mb-1">
                Email Address
              </span>
              <div className="flex items-center gap-2">
                <span className="font-mono text-neutral-800 dark:text-neutral-200">
                  {user?.email || "—"}
                </span>
                {user?.is_verified && (
                  <span
                    className="inline-flex items-center gap-1 text-[10px] text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40 px-1.5 py-0.5 rounded font-medium"
                    title="Email verified"
                  >
                    <CheckCircle2 className="w-3 h-3" />
                    Verified
                  </span>
                )}
              </div>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px] mb-1">
                Account Role
              </span>
              <span className="inline-block px-2 py-0.5 rounded text-[11px] font-mono font-semibold bg-neutral-100 dark:bg-neutral-800 text-neutral-700 dark:text-neutral-300 border border-neutral-200 dark:border-neutral-700">
                {user?.role || "USER"}
              </span>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px] mb-1">
                Member Since
              </span>
              <p className="font-mono text-neutral-800 dark:text-neutral-200">
                {user?.created_at
                  ? new Date(user.created_at).toLocaleDateString()
                  : "—"}
              </p>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px] mb-1">
                Last Login
              </span>
              <p className="font-mono text-neutral-800 dark:text-neutral-200">
                {user?.last_login_at
                  ? new Date(user.last_login_at).toLocaleString()
                  : "Current session"}
              </p>
            </div>
          </div>
        </div>

        {/* Appearance Card */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-5 shadow-2xs">
          <div className="flex items-center gap-2 pb-3.5 border-b border-neutral-100 dark:border-neutral-800/80 mb-4">
            <Palette className="w-4 h-4 text-emerald-500" />
            <h2 className="text-xs font-bold uppercase tracking-wider text-neutral-500">
              Appearance & Theme
            </h2>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <button
              type="button"
              onClick={() => setTheme("dark")}
              className={`p-3.5 rounded-xl border text-left flex items-start justify-between transition-colors cursor-pointer ${
                theme === "dark"
                  ? "border-emerald-500 bg-emerald-50/40 dark:bg-emerald-950/20"
                  : "border-neutral-200 dark:border-neutral-800 hover:bg-neutral-50 dark:hover:bg-neutral-900/40"
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Moon className="w-4 h-4 text-neutral-700 dark:text-neutral-300" />
                <div>
                  <p className="text-xs font-bold text-neutral-900 dark:text-white">
                    Dark Theme
                  </p>
                  <p className="text-[11px] text-neutral-500">
                    High contrast low-light palette
                  </p>
                </div>
              </div>
              {theme === "dark" && (
                <Check className="w-4 h-4 text-emerald-500 shrink-0" />
              )}
            </button>

            <button
              type="button"
              onClick={() => setTheme("light")}
              className={`p-3.5 rounded-xl border text-left flex items-start justify-between transition-colors cursor-pointer ${
                theme === "light"
                  ? "border-emerald-500 bg-emerald-50/40 dark:bg-emerald-950/20"
                  : "border-neutral-200 dark:border-neutral-800 hover:bg-neutral-50 dark:hover:bg-neutral-900/40"
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Sun className="w-4 h-4 text-amber-500" />
                <div>
                  <p className="text-xs font-bold text-neutral-900 dark:text-white">
                    Light Theme
                  </p>
                  <p className="text-[11px] text-neutral-500">
                    Clean crisp daylight palette
                  </p>
                </div>
              </div>
              {theme === "light" && (
                <Check className="w-4 h-4 text-emerald-500 shrink-0" />
              )}
            </button>

            <button
              type="button"
              onClick={() => setTheme("system")}
              className={`p-3.5 rounded-xl border text-left flex items-start justify-between transition-colors cursor-pointer ${
                theme === "system"
                  ? "border-emerald-500 bg-emerald-50/40 dark:bg-emerald-950/20"
                  : "border-neutral-200 dark:border-neutral-800 hover:bg-neutral-50 dark:hover:bg-neutral-900/40"
              }`}
            >
              <div className="flex items-center gap-2.5">
                <MonitorIcon className="w-4 h-4 text-neutral-700 dark:text-neutral-300" />
                <div>
                  <p className="text-xs font-bold text-neutral-900 dark:text-white">
                    System Default
                  </p>
                  <p className="text-[11px] text-neutral-500">
                    Sync with OS preference
                  </p>
                </div>
              </div>
              {theme === "system" && (
                <Check className="w-4 h-4 text-emerald-500 shrink-0" />
              )}
            </button>
          </div>
        </div>

        {/* Security & Backend Connection Card */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-5 shadow-2xs">
          <div className="flex items-center gap-2 pb-3.5 border-b border-neutral-100 dark:border-neutral-800/80 mb-4">
            <Server className="w-4 h-4 text-emerald-500" />
            <h2 className="text-xs font-bold uppercase tracking-wider text-neutral-500">
              Backend Integration & Security
            </h2>
          </div>

          <div className="space-y-3 text-xs">
            <div className="flex items-center justify-between py-1 border-b border-neutral-100 dark:border-neutral-800/60">
              <span className="text-neutral-500">API Gateway Endpoint</span>
              <span className="font-mono font-medium text-neutral-900 dark:text-white">
                {apiBaseUrl}
              </span>
            </div>

            <div className="flex items-center justify-between py-1 border-b border-neutral-100 dark:border-neutral-800/60">
              <span className="text-neutral-500">Authentication Protocol</span>
              <span className="font-mono text-neutral-800 dark:text-neutral-200">
                JWT Bearer (15m in memory) + HttpOnly Refresh Cookie
              </span>
            </div>

            <div className="flex items-center justify-between py-1 border-b border-neutral-100 dark:border-neutral-800/60">
              <span className="text-neutral-500">Local Storage Secrets</span>
              <span className="text-emerald-600 dark:text-emerald-400 font-medium">
                None (Zero credentials stored in localStorage)
              </span>
            </div>

            <div className="flex items-center justify-between py-1">
              <span className="text-neutral-500">Session State</span>
              <span className="inline-flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400 font-semibold">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                Active Authenticated Session
              </span>
            </div>
          </div>
        </div>

        {/* Sign Out Card */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-5 shadow-2xs flex items-center justify-between">
          <div>
            <h3 className="text-xs font-bold text-neutral-900 dark:text-white">
              Terminate Active Session
            </h3>
            <p className="text-[11px] text-neutral-500 mt-0.5">
              Revokes your refresh token family and clears in-memory credentials.
            </p>
          </div>

          <Button
            variant="danger"
            size="sm"
            onClick={handleLogout}
            isLoading={loggingOut}
          >
            <LogOut className="w-3.5 h-3.5 mr-1" />
            Sign Out
          </Button>
        </div>
      </div>
    </DashboardLayout>
  );
}
