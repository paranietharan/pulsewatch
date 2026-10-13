import React from "react";
import Link from "next/link";
import { PulseLogoIcon } from "@/components/icons";
import { ThemeToggle } from "@/components/theme/ThemeToggle";

export function Navbar() {
  return (
    <header className="sticky top-0 z-50 backdrop-blur-md bg-white/85 dark:bg-[#07090e]/85 border-b border-neutral-200/80 dark:border-neutral-800/80 transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2.5 group">
          <div className="w-7 h-7 rounded-lg bg-neutral-100 dark:bg-[#0d121f] border border-neutral-200 dark:border-neutral-800 flex items-center justify-center p-0.5 group-hover:border-emerald-500/50 transition-colors">
            <PulseLogoIcon className="w-full h-full" />
          </div>
          <div className="flex items-center gap-2">
            <span className="text-sm font-bold tracking-tight text-neutral-900 dark:text-white font-[family-name:var(--font-heading)]">
              PulseWatch
            </span>
            <span className="text-[10px] font-mono px-1.5 py-0.2 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-600 dark:text-neutral-400 font-medium border border-neutral-200 dark:border-neutral-700">
              v2.0
            </span>
          </div>
        </Link>

        <nav className="hidden md:flex items-center gap-7 text-xs font-medium text-neutral-600 dark:text-neutral-400">
          <a href="#features" className="hover:text-neutral-900 dark:hover:text-white transition-colors">
            Features
          </a>
          <a href="#pricing" className="hover:text-neutral-900 dark:hover:text-white transition-colors">
            Pricing
          </a>
        </nav>

        <div className="flex items-center gap-2.5">
          <div className="hidden lg:flex items-center gap-2 px-2.5 py-1 rounded-full bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-500/20 text-[11px] font-medium text-emerald-700 dark:text-emerald-400">
            <span className="radar-beacon" />
            <span>14 Nodes Active</span>
          </div>

          <a
            href="http://localhost:8080/health"
            target="_blank"
            rel="noreferrer"
            className="text-xs font-medium text-neutral-600 hover:text-neutral-950 dark:text-neutral-400 dark:hover:text-white px-2 py-1 transition-colors hidden sm:block"
          >
            API Status
          </a>

          <ThemeToggle />

          <a
            href="#pricing"
            className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg font-semibold text-xs text-white bg-neutral-900 hover:bg-neutral-800 dark:bg-white dark:text-neutral-900 dark:hover:bg-neutral-100 transition-colors shadow-sm"
          >
            Start Monitoring
          </a>
        </div>
      </div>
    </header>
  );
}
