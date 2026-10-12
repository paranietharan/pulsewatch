"use client";

import React, { useState, useEffect } from "react";
import { MONITORS } from "@/lib/constants/landing-data";
import { MonitorProfile } from "@/types/landing";
import { MonitorList } from "./MonitorList";
import { TelemetryViewer } from "./TelemetryViewer";

export function LiveConsole() {
  const [selectedMonitor, setSelectedMonitor] = useState<MonitorProfile>(MONITORS[0]);
  const [tickerSecond, setTickerSecond] = useState(4);

  useEffect(() => {
    const timer = setInterval(() => {
      setTickerSecond((prev) => (prev <= 1 ? 5 : prev - 1));
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  return (
    <section id="demo" className="py-10 relative z-10 scroll-mt-20">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="bg-white dark:bg-[#0a0e1a] p-4 sm:p-6 rounded-2xl border border-neutral-200 dark:border-neutral-800 shadow-sm dark:shadow-2xl relative overflow-hidden transition-colors">
          {/* Console Header Bar */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-neutral-200/80 dark:border-neutral-800/80 gap-3">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-rose-500/80 inline-block" />
                <span className="w-2.5 h-2.5 rounded-full bg-amber-500/80 inline-block" />
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-500/80 inline-block" />
              </div>
              <div className="h-3.5 w-[1px] bg-neutral-200 dark:bg-neutral-800 mx-0.5" />
              <span className="text-xs font-mono text-neutral-500 dark:text-neutral-400">
                pulsewatch://cluster.prod/monitors
              </span>
            </div>

            <div className="flex items-center gap-3 text-xs">
              <span className="flex items-center gap-1.5 text-neutral-500 dark:text-neutral-400 font-mono text-[11px]">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                Next probe in <strong className="text-neutral-900 dark:text-white font-mono">{tickerSecond}s</strong>
              </span>
              <span className="px-2 py-0.5 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-700 dark:text-neutral-300 font-mono text-[11px] border border-neutral-200 dark:border-neutral-700">
                Region: {selectedMonitor.region}
              </span>
            </div>
          </div>

          {/* Console Dual Pane */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-5 pt-5">
            <MonitorList
              monitors={MONITORS}
              selectedMonitor={selectedMonitor}
              onSelect={setSelectedMonitor}
            />

            <TelemetryViewer monitor={selectedMonitor} />
          </div>
        </div>
      </div>
    </section>
  );
}
