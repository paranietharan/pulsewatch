import React from "react";
import { Incident } from "@/types/api";
import { Badge } from "@/components/ui/Badge";

export function IncidentHistoryTable({
  incidents,
  showMonitorName = false,
}: {
  incidents: Incident[];
  showMonitorName?: boolean;
}) {
  if (!incidents || incidents.length === 0) {
    return (
      <div className="py-8 text-center text-xs text-neutral-500">
        No incidents reported. All systems operational.
      </div>
    );
  }

  const formatDuration = (sec?: number | null) => {
    if (!sec && sec !== 0) return "Ongoing";
    if (sec < 60) return `${sec}s`;
    if (sec < 3600) return `${Math.floor(sec / 60)}m ${sec % 60}s`;
    return `${Math.floor(sec / 3600)}h ${Math.floor((sec % 3600) / 60)}m`;
  };

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left border-collapse text-xs">
        <thead>
          <tr className="border-b border-neutral-200 dark:border-neutral-800 text-[11px] font-semibold text-neutral-500 uppercase tracking-wider">
            <th className="py-2.5 px-3">Status</th>
            {showMonitorName && <th className="py-2.5 px-3">Monitor</th>}
            <th className="py-2.5 px-3">Cause Summary</th>
            <th className="py-2.5 px-3">Started At</th>
            <th className="py-2.5 px-3">Resolved At</th>
            <th className="py-2.5 px-3">Duration</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-neutral-100 dark:divide-neutral-800/60">
          {incidents.map((inc) => (
            <tr
              key={inc.id}
              className="hover:bg-neutral-50/70 dark:hover:bg-neutral-900/40 transition-colors"
            >
              <td className="py-2.5 px-3 whitespace-nowrap">
                <Badge status={inc.status} />
              </td>
              {showMonitorName && (
                <td className="py-2.5 px-3 font-medium text-neutral-900 dark:text-white">
                  {inc.monitor_name || inc.monitor_id}
                </td>
              )}
              <td className="py-2.5 px-3 text-neutral-700 dark:text-neutral-300">
                {inc.cause_summary || "Automated threshold failure detected"}
              </td>
              <td className="py-2.5 px-3 font-mono text-[11px] text-neutral-500 whitespace-nowrap">
                {new Date(inc.started_at).toLocaleString()}
              </td>
              <td className="py-2.5 px-3 font-mono text-[11px] text-neutral-500 whitespace-nowrap">
                {inc.resolved_at ? new Date(inc.resolved_at).toLocaleString() : "—"}
              </td>
              <td className="py-2.5 px-3 font-mono text-[11px]">
                <span
                  className={
                    inc.status === "OPEN"
                      ? "text-rose-600 dark:text-rose-400 font-semibold"
                      : "text-neutral-600 dark:text-neutral-400"
                  }
                >
                  {formatDuration(inc.duration_seconds)}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
