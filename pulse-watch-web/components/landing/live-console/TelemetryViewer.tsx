import React from "react";
import { MonitorProfile } from "@/types/landing";

interface TelemetryViewerProps {
  monitor: MonitorProfile;
}

export function TelemetryViewer({ monitor }: TelemetryViewerProps) {
  return (
    <div className="lg:col-span-7 bg-neutral-50/70 dark:bg-[#0c101d]/70 rounded-xl p-4 sm:p-5 border border-neutral-200/80 dark:border-neutral-800/80 flex flex-col gap-5 transition-colors">
      {/* Header details */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-3 border-b border-neutral-200/80 dark:border-neutral-800/80">
        <div>
          <h3 className="text-sm font-bold text-neutral-900 dark:text-white flex items-center gap-2 font-[family-name:var(--font-heading)]">
            {monitor.name}
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-500/25">
              {monitor.status} 100%
            </span>
          </h3>
          <p className="text-[11px] font-mono text-neutral-500 dark:text-neutral-400 mt-0.5 truncate max-w-sm">
            {monitor.endpoint}
          </p>
        </div>

        <div className="flex items-center gap-2 text-[11px] font-mono">
          <span className="px-2 py-0.5 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-700 dark:text-neutral-300 border border-neutral-200 dark:border-neutral-700">
            SSL Valid ({monitor.sslDays}d)
          </span>
          <span className="px-2 py-0.5 rounded bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-500/25">
            SLA: {monitor.uptime90d}
          </span>
        </div>
      </div>

      {/* Latency Waterfall Decomposition */}
      <div>
        <div className="flex items-center justify-between text-xs mb-2">
          <span className="font-mono text-[11px] uppercase tracking-wider text-neutral-500 dark:text-neutral-400 font-semibold">
            Round-Trip Latency Decomposition
          </span>
          <span className="font-mono text-xs font-bold text-neutral-900 dark:text-white">
            Total: {monitor.latency}ms
          </span>
        </div>

        {/* Multi-segment bar */}
        <div className="h-2.5 w-full bg-neutral-200 dark:bg-neutral-800 rounded-full overflow-hidden flex">
          <div
            style={{ width: `${(monitor.dns / monitor.latency) * 100}%` }}
            className="bg-sky-500 transition-all duration-300"
            title={`DNS: ${monitor.dns}ms`}
          />
          <div
            style={{ width: `${(monitor.tcp / monitor.latency) * 100}%` }}
            className="bg-cyan-500 transition-all duration-300"
            title={`TCP: ${monitor.tcp}ms`}
          />
          <div
            style={{ width: `${(monitor.tls / monitor.latency) * 100}%` }}
            className="bg-indigo-500 transition-all duration-300"
            title={`TLS: ${monitor.tls}ms`}
          />
          <div
            style={{ width: `${(monitor.ttfb / monitor.latency) * 100}%` }}
            className="bg-emerald-500 transition-all duration-300"
            title={`TTFB: ${monitor.ttfb}ms`}
          />
        </div>

        {/* Minimal Legend */}
        <div className="grid grid-cols-4 gap-2 mt-2 text-[10px] font-mono text-neutral-600 dark:text-neutral-400 text-center">
          <div className="bg-white dark:bg-[#070b14] py-1 rounded border border-neutral-200/70 dark:border-neutral-800">
            <span className="inline-block w-1.5 h-1.5 rounded-full bg-sky-500 mr-1" />
            DNS: {monitor.dns}ms
          </div>
          <div className="bg-white dark:bg-[#070b14] py-1 rounded border border-neutral-200/70 dark:border-neutral-800">
            <span className="inline-block w-1.5 h-1.5 rounded-full bg-cyan-500 mr-1" />
            TCP: {monitor.tcp}ms
          </div>
          <div className="bg-white dark:bg-[#070b14] py-1 rounded border border-neutral-200/70 dark:border-neutral-800">
            <span className="inline-block w-1.5 h-1.5 rounded-full bg-indigo-500 mr-1" />
            TLS: {monitor.tls}ms
          </div>
          <div className="bg-white dark:bg-[#070b14] py-1 rounded border border-neutral-200/70 dark:border-neutral-800">
            <span className="inline-block w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1" />
            TTFB: {monitor.ttfb}ms
          </div>
        </div>
      </div>

      {/* 90-Day Historical Uptime Strip */}
      <div>
        <div className="flex items-center justify-between text-xs mb-1.5">
          <span className="font-mono text-[11px] uppercase tracking-wider text-neutral-500 dark:text-neutral-400 font-semibold">
            90-Day Historical Uptime
          </span>
          <span className="font-mono text-[11px] text-emerald-600 dark:text-emerald-400 font-semibold">
            {monitor.uptime90d} Available
          </span>
        </div>

        <div className="flex items-center gap-[2px] h-5 w-full py-0.5">
          {Array.from({ length: 50 }).map((_, idx) => (
            <div
              key={idx}
              className="uptime-segment bg-emerald-500/80 hover:bg-emerald-500 cursor-pointer"
              title={`Day ${90 - idx}: 100% (Avg ${monitor.latency}ms)`}
            />
          ))}
        </div>

        <div className="flex justify-between text-[10px] font-mono text-neutral-400 dark:text-neutral-500 mt-1">
          <span>90 days ago</span>
          <span>Continuous Probing</span>
          <span>Today (Operational)</span>
        </div>
      </div>

      {/* Regional Probing Latency Grid */}
      <div>
        <div className="text-[11px] font-mono uppercase tracking-wider text-neutral-500 dark:text-neutral-400 font-semibold mb-2">
          Multi-Region Edge Telemetry
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          {monitor.pings.map((ping, idx) => (
            <div
              key={idx}
              className="bg-white dark:bg-[#070b14] p-2 rounded-lg border border-neutral-200/70 dark:border-neutral-800 flex flex-col"
            >
              <span className="text-[10px] text-neutral-500 dark:text-neutral-400 truncate">{ping.loc}</span>
              <div className="flex items-center justify-between mt-1">
                <span className="text-xs font-bold font-mono text-neutral-900 dark:text-white">{ping.ms}ms</span>
                <span
                  className={`w-1.5 h-1.5 rounded-full ${
                    ping.optimal ? "bg-emerald-500" : "bg-sky-500"
                  }`}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
