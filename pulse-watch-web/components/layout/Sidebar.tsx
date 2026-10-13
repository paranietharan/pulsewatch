"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Activity,
  AlertTriangle,
  Bell,
  Settings,
  LogOut,
  X,
} from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import { PulseLogoIcon } from "@/components/icons";

const NAV_ITEMS = [
  { label: "Overview", href: "/dashboard", icon: LayoutDashboard },
  { label: "Monitors", href: "/monitors", icon: Activity },
  { label: "Incidents", href: "/incidents", icon: AlertTriangle },
  { label: "Alert Channels", href: "/channels", icon: Bell },
  { label: "Settings", href: "/settings", icon: Settings },
];

export function Sidebar({
  isOpen,
  onClose,
}: {
  isOpen: boolean;
  onClose: () => void;
}) {
  const pathname = usePathname();
  const { user, logout } = useAuth();

  return (
    <>
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/40 dark:bg-black/60 z-40 lg:hidden backdrop-blur-xs"
          onClick={onClose}
        />
      )}

      <aside
        className={`fixed top-0 bottom-0 left-0 z-50 w-60 bg-white dark:bg-[#07090e] border-r border-neutral-200 dark:border-neutral-800/80 flex flex-col justify-between transition-transform duration-200 lg:translate-x-0 ${
          isOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div>
          <div className="h-14 px-4 flex items-center justify-between border-b border-neutral-100 dark:border-neutral-800/80">
            <Link href="/dashboard" className="flex items-center gap-2.5">
              <div className="w-6 h-6 rounded-md bg-neutral-100 dark:bg-[#0e1424] border border-neutral-200 dark:border-neutral-800 flex items-center justify-center p-0.5">
                <PulseLogoIcon className="w-full h-full" />
              </div>
              <span className="text-xs font-bold tracking-tight text-neutral-900 dark:text-white">
                PulseWatch
              </span>
              <span className="text-[10px] font-mono px-1 py-0.2 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-500 font-medium">
                v2.0
              </span>
            </Link>
            <button
              onClick={onClose}
              className="lg:hidden text-neutral-400 hover:text-neutral-700 dark:hover:text-white p-1"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          <nav className="p-3 flex flex-col gap-1">
            {NAV_ITEMS.map((item) => {
              const Icon = item.icon;
              const isActive =
                item.href === "/dashboard"
                  ? pathname === "/dashboard"
                  : pathname.startsWith(item.href);

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={onClose}
                  className={`flex items-center gap-2.5 px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
                    isActive
                      ? "bg-neutral-100 text-neutral-900 dark:bg-neutral-800/80 dark:text-white font-semibold"
                      : "text-neutral-600 hover:text-neutral-900 hover:bg-neutral-50 dark:text-neutral-400 dark:hover:text-neutral-200 dark:hover:bg-neutral-900/50"
                  }`}
                >
                  <Icon className={`w-4 h-4 ${isActive ? "text-emerald-500" : "text-neutral-400"}`} />
                  <span>{item.label}</span>
                </Link>
              );
            })}
          </nav>
        </div>

        <div className="p-3 border-t border-neutral-100 dark:border-neutral-800/80">
          <div className="flex items-center justify-between px-2.5 py-2 mb-2 bg-neutral-50 dark:bg-neutral-900/40 rounded-lg border border-neutral-100 dark:border-neutral-800/60">
            <div className="min-w-0 pr-2">
              <p className="text-[11px] font-semibold text-neutral-900 dark:text-white truncate">
                {user ? `${user.first_name} ${user.last_name}` : "Developer"}
              </p>
              <p className="text-[10px] text-neutral-500 truncate">
                {user?.email || ""}
              </p>
            </div>
            <button
              onClick={() => logout()}
              title="Sign out"
              className="text-neutral-400 hover:text-rose-600 dark:hover:text-rose-400 transition-colors p-1"
            >
              <LogOut className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </aside>
    </>
  );
}
