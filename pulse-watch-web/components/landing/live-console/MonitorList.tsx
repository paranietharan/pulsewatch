import React from "react";
import { MonitorProfile } from "@/types/landing";

interface MonitorListProps {
  monitors: MonitorProfile[];
  selectedMonitor: MonitorProfile;
  onSelect: (monitor: MonitorProfile) => void;
}

export function MonitorList({
  monitors,
  selectedMonitor,
  onSelect,
}: MonitorListProps) {
  return (
    <div className="lg:col-span-5 flex flex-col gap-2">
      <div className="text-[11px] font-mono uppercase font-semibold text-neutral-500 dark:text-neutral-400 tracking-wider mb-0.5 px-1">
        Active Synthetic Endpoints ({monitors.length})
      </div>

      {monitors.map((mon) => {
        const isSelected = selectedMonitor.id === mon.id;
        return (
          <button
            key={mon.id}
            onClick={() => onSelect(mon)}
            className={`text-left p-3 rounded-xl border transition-all flex flex-col gap-1 cursor-pointer ${
              isSelected
                ? "bg-neutral-100 dark:bg-[#131b2e] border-neutral-300 dark:border-neutral-700 shadow-sm ring-1 ring-neutral-300 dark:ring-neutral-700"
                : "bg-neutral-50/60 dark:bg-[#090d18]/50 border-neutral-200/80 dark:border-neutral-800/80 hover:bg-neutral-100/70 dark:hover:bg-[#0f1526]"
            }`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span
                  className={`w-2 h-2 rounded-full ${
                    mon.status === "UP" ? "bg-emerald-500" : "bg-rose-500"
                  }`}
                />
                <span className="text-xs font-semibold text-neutral-900 dark:text-white truncate max-w-[170px]">
                  {mon.name}
                </span>
              </div>
              <div className="flex items-center gap-1.5 font-mono text-[11px]">
                <span className="text-emerald-600 dark:text-emerald-400 font-medium">
                  {mon.method} {mon.httpCode}
                </span>
                <span className="text-neutral-500 dark:text-neutral-400">{mon.latency}ms</span>
              </div>
            </div>

            <div className="text-[11px] font-mono text-neutral-500 dark:text-neutral-400 truncate pl-4">
              {mon.endpoint}
            </div>
          </button>
        );
      })}
    </div>
  );
}
