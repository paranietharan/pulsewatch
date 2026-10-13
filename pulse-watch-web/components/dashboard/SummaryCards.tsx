import React from "react";
import { DashboardSummary } from "@/types/api";
import { Activity, CheckCircle, AlertOctagon, AlertTriangle, PauseCircle, Clock, Percent } from "lucide-react";

interface SummaryCardsProps {
  summary: DashboardSummary | null;
  isLoading: boolean;
}

export function SummaryCards({ summary, isLoading }: SummaryCardsProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <div
            key={i}
            className="h-24 bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800 rounded-xl animate-pulse"
          />
        ))}
      </div>
    );
  }

  const s = summary || {
    total_monitors: 0,
    healthy_monitors: 0,
    degraded_monitors: 0,
    down_monitors: 0,
    paused_monitors: 0,
    active_incidents: 0,
    avg_uptime_24h: 100,
    avg_response_time_ms: 0,
  };

  const cards = [
    {
      label: "Total Monitors",
      value: s.total_monitors,
      icon: Activity,
      color: "text-neutral-700 dark:text-neutral-300",
      bg: "bg-neutral-100 dark:bg-neutral-800",
    },
    {
      label: "Healthy",
      value: s.healthy_monitors,
      icon: CheckCircle,
      color: "text-emerald-600 dark:text-emerald-400",
      bg: "bg-emerald-50 dark:bg-emerald-950/40",
    },
    {
      label: "Down / Outages",
      value: s.down_monitors,
      icon: AlertOctagon,
      color: s.down_monitors > 0 ? "text-rose-600 dark:text-rose-400" : "text-neutral-500",
      bg: s.down_monitors > 0 ? "bg-rose-50 dark:bg-rose-950/40" : "bg-neutral-100 dark:bg-neutral-800",
    },
    {
      label: "Active Incidents",
      value: s.active_incidents,
      icon: AlertTriangle,
      color: s.active_incidents > 0 ? "text-rose-600 dark:text-rose-400 font-bold" : "text-neutral-500",
      bg: s.active_incidents > 0 ? "bg-rose-50 dark:bg-rose-950/40" : "bg-neutral-100 dark:bg-neutral-800",
    },
    {
      label: "Paused",
      value: s.paused_monitors,
      icon: PauseCircle,
      color: "text-neutral-500",
      bg: "bg-neutral-100 dark:bg-neutral-800",
    },
    {
      label: "Degraded",
      value: s.degraded_monitors,
      icon: AlertTriangle,
      color: s.degraded_monitors > 0 ? "text-amber-600 dark:text-amber-400" : "text-neutral-500",
      bg: s.degraded_monitors > 0 ? "bg-amber-50 dark:bg-amber-950/40" : "bg-neutral-100 dark:bg-neutral-800",
    },
    {
      label: "Avg 24h Uptime",
      value: `${s.avg_uptime_24h ? s.avg_uptime_24h.toFixed(2) : "100"}%`,
      icon: Percent,
      color: "text-neutral-900 dark:text-white",
      bg: "bg-neutral-100 dark:bg-neutral-800",
    },
    {
      label: "Avg Response Time",
      value: `${s.avg_response_time_ms || 0}ms`,
      icon: Clock,
      color: "text-neutral-900 dark:text-white",
      bg: "bg-neutral-100 dark:bg-neutral-800",
    },
  ];

  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
      {cards.map((c, idx) => {
        const Icon = c.icon;
        return (
          <div
            key={idx}
            className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 flex flex-col justify-between transition-colors shadow-2xs"
          >
            <div className="flex items-center justify-between text-neutral-500">
              <span className="text-[11px] font-medium text-neutral-500 dark:text-neutral-400">
                {c.label}
              </span>
              <div className={`w-6 h-6 rounded-md ${c.bg} ${c.color} flex items-center justify-center`}>
                <Icon className="w-3.5 h-3.5" />
              </div>
            </div>
            <div className={`text-xl font-bold font-mono tracking-tight mt-2 ${c.color}`}>
              {c.value}
            </div>
          </div>
        );
      })}
    </div>
  );
}
