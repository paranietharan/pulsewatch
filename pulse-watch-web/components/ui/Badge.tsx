import React from "react";
import { MonitorStatus, IncidentStatus } from "@/types/api";

type BadgeStatus = MonitorStatus | IncidentStatus | "ACTIVE" | "PAUSED" | "SUCCESS" | "WARNING" | "DANGER" | "INFO";

interface BadgeProps {
  status: BadgeStatus | string;
  className?: string;
  showDot?: boolean;
}

export function Badge({ status, className = "", showDot = true }: BadgeProps) {
  const normalized = status.toUpperCase();

  let styles = "bg-neutral-100 text-neutral-700 border-neutral-300 dark:bg-neutral-800 dark:text-neutral-300 dark:border-neutral-700";
  let dotColor = "bg-neutral-400";

  if (normalized === "UP" || normalized === "RESOLVED" || normalized === "ACTIVE" || normalized === "SUCCESS") {
    styles = "bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/30 dark:text-emerald-400 dark:border-emerald-800/40";
    dotColor = "bg-emerald-500";
  } else if (normalized === "DOWN" || normalized === "OPEN" || normalized === "DANGER") {
    styles = "bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/30 dark:text-rose-400 dark:border-rose-800/40";
    dotColor = "bg-rose-500";
  } else if (normalized === "DEGRADED" || normalized === "ACKNOWLEDGED" || normalized === "WARNING") {
    styles = "bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/30 dark:text-amber-400 dark:border-amber-800/40";
    dotColor = "bg-amber-500";
  } else if (normalized === "PAUSED" || normalized === "PENDING") {
    styles = "bg-neutral-100 text-neutral-600 border-neutral-200 dark:bg-neutral-800/60 dark:text-neutral-400 dark:border-neutral-700";
    dotColor = "bg-neutral-400";
  }

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[11px] font-mono font-medium border ${styles} ${className}`}
    >
      {showDot && <span className={`w-1.5 h-1.5 rounded-full ${dotColor}`} />}
      {normalized}
    </span>
  );
}
