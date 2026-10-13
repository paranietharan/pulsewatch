import React from "react";
import { Heartbeat } from "@/types/api";
import { Badge } from "@/components/ui/Badge";

export function HeartbeatResultsTable({ results }: { results: Heartbeat[] }) {
  if (!results || results.length === 0) {
    return (
      <div className="py-8 text-center text-xs text-neutral-500">
        No probe results recorded yet.
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left border-collapse text-xs">
        <thead>
          <tr className="border-b border-neutral-200 dark:border-neutral-800 text-[11px] font-semibold text-neutral-500 uppercase tracking-wider">
            <th className="py-2.5 px-3">Status</th>
            <th className="py-2.5 px-3">HTTP Code</th>
            <th className="py-2.5 px-3">Response Time</th>
            <th className="py-2.5 px-3">DNS / TCP / TLS</th>
            <th className="py-2.5 px-3">Region</th>
            <th className="py-2.5 px-3">Checked At</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-neutral-100 dark:divide-neutral-800/60">
          {results.map((r) => (
            <tr
              key={r.id}
              className="hover:bg-neutral-50/70 dark:hover:bg-neutral-900/40 transition-colors"
            >
              <td className="py-2.5 px-3 whitespace-nowrap">
                <Badge status={r.status} />
              </td>
              <td className="py-2.5 px-3 font-mono font-medium">
                {r.status_code || "—"}
              </td>
              <td className="py-2.5 px-3 font-mono">
                <span
                  className={
                    r.response_time_ms > 1000
                      ? "text-rose-600 dark:text-rose-400 font-semibold"
                      : r.response_time_ms > 300
                      ? "text-amber-600 dark:text-amber-400"
                      : "text-emerald-600 dark:text-emerald-400 font-medium"
                  }
                >
                  {r.response_time_ms}ms
                </span>
                {r.error_message && (
                  <p className="text-[10px] text-rose-500 mt-0.5 truncate max-w-xs" title={r.error_message}>
                    {r.error_message}
                  </p>
                )}
              </td>
              <td className="py-2.5 px-3 font-mono text-[11px] text-neutral-500">
                {r.dns_lookup_ms}ms / {r.tcp_connect_ms}ms / {r.tls_handshake_ms}ms
              </td>
              <td className="py-2.5 px-3 text-neutral-500 font-mono text-[11px]">
                {r.region || "default"}
              </td>
              <td className="py-2.5 px-3 text-neutral-500 font-mono text-[11px] whitespace-nowrap">
                {new Date(r.checked_at).toLocaleString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
