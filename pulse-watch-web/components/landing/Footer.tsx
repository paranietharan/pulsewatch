import React from "react";
import Link from "next/link";
import { PulseLogoIcon } from "@/components/icons";

export function Footer() {
  return (
    <footer className="border-t border-neutral-200 dark:border-neutral-800/80 bg-white dark:bg-[#07090e] py-8 text-xs text-neutral-500 relative z-10 transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2.5">
          <PulseLogoIcon className="w-4 h-4 text-emerald-500" />
          <span className="text-neutral-900 dark:text-neutral-200 font-semibold tracking-tight">
            PulseWatch
          </span>
          <span className="text-neutral-300 dark:text-neutral-700">|</span>
          <span className="text-neutral-500 dark:text-neutral-400">
            &copy; 2026 PulseWatch. Distributed Uptime Monitoring.
          </span>
        </div>

        <div className="flex items-center gap-5 text-neutral-500 dark:text-neutral-400 text-xs">
          <a
            href="#features"
            className="hover:text-neutral-900 dark:hover:text-white transition-colors"
          >
            Features
          </a>
          <a
            href="#pricing"
            className="hover:text-neutral-900 dark:hover:text-white transition-colors"
          >
            Pricing
          </a>
          <Link
            href="/login"
            className="hover:text-neutral-900 dark:hover:text-white transition-colors"
          >
            Sign In
          </Link>
          <Link
            href="/signup"
            className="hover:text-neutral-900 dark:hover:text-white transition-colors"
          >
            Sign Up
          </Link>
        </div>
      </div>
    </footer>
  );
}
