"use client";

import React, { useState } from "react";
import { Heartbeat } from "@/types/api";

interface LatencyChartProps {
  results: Heartbeat[];
}

export function LatencyChart({ results }: LatencyChartProps) {
  const [hovered, setHovered] = useState<Heartbeat | null>(null);

  if (!results || results.length === 0) {
    return (
      <div className="h-40 flex items-center justify-center text-xs text-neutral-400">
        No latency data recorded yet
      </div>
    );
  }

  const sorted = [...results].reverse();
  const maxMs = Math.max(...sorted.map((r) => r.response_time_ms), 100);
  const minMs = Math.min(...sorted.map((r) => r.response_time_ms));
  const avgMs = Math.round(
    sorted.reduce((acc, r) => acc + r.response_time_ms, 0) / sorted.length
  );

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between text-xs text-neutral-500">
        <div className="flex items-center gap-4">
          <span>
            Avg: <strong className="text-neutral-900 dark:text-white font-mono">{avgMs}ms</strong>
          </span>
          <span>
            Min: <strong className="text-neutral-900 dark:text-white font-mono">{minMs}ms</strong>
          </span>
          <span>
            Max: <strong className="text-neutral-900 dark:text-white font-mono">{maxMs}ms</strong>
          </span>
        </div>

        {hovered && (
          <div className="text-[11px] font-mono text-neutral-700 dark:text-neutral-300">
            {hovered.response_time_ms}ms · {hovered.status} ·{" "}
            {new Date(hovered.checked_at).toLocaleTimeString()}
          </div>
        )}
      </div>

      <div className="h-28 w-full flex items-end gap-1 pt-4 pb-1 border-b border-neutral-200 dark:border-neutral-800">
        {sorted.map((item, idx) => {
          const heightPercent = Math.max((item.response_time_ms / maxMs) * 100, 6);
          const isUp = item.status === "UP";
          const isDown = item.status === "DOWN";

          let barColor = "bg-neutral-300 dark:bg-neutral-700";
          if (isUp) barColor = "bg-emerald-500 hover:bg-emerald-400";
          else if (isDown) barColor = "bg-rose-500 hover:bg-rose-400";
          else barColor = "bg-amber-500 hover:bg-amber-400";

          return (
            <div
              key={idx}
              className="flex-1 h-full flex items-end group cursor-pointer relative"
              onMouseEnter={() => setHovered(item)}
              onMouseLeave={() => setHovered(null)}
            >
              <div
                style={{ height: `${heightPercent}%` }}
                className={`w-full rounded-xs transition-all ${barColor}`}
              />
            </div>
          );
        })}
      </div>

      <div className="flex justify-between text-[10px] text-neutral-400 font-mono">
        <span>Older</span>
        <span>Latest Checks ({sorted.length})</span>
      </div>
    </div>
  );
}
