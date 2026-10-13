"use client";

import React from "react";
import { Menu } from "lucide-react";
import { ThemeToggle } from "@/components/theme/ThemeToggle";

interface HeaderProps {
  title: string;
  description?: string;
  action?: React.ReactNode;
  onOpenSidebar: () => void;
}

export function Header({
  title,
  description,
  action,
  onOpenSidebar,
}: HeaderProps) {
  return (
    <header className="sticky top-0 z-30 h-14 bg-white/80 dark:bg-[#07090e]/80 backdrop-blur-md border-b border-neutral-200/80 dark:border-neutral-800/80 flex items-center justify-between px-4 sm:px-6">
      <div className="flex items-center gap-3">
        <button
          onClick={onOpenSidebar}
          className="lg:hidden text-neutral-500 hover:text-neutral-800 dark:hover:text-white p-1"
          aria-label="Open sidebar"
        >
          <Menu className="w-5 h-5" />
        </button>
        <div>
          <h1 className="text-sm font-bold tracking-tight text-neutral-900 dark:text-white">
            {title}
          </h1>
          {description && (
            <p className="hidden sm:block text-[11px] text-neutral-500 dark:text-neutral-400">
              {description}
            </p>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2.5">
        <ThemeToggle />
        {action}
      </div>
    </header>
  );
}
