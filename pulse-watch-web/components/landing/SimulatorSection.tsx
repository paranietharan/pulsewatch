"use client";

import React, { useState } from "react";
import { SIMULATOR_STAGES } from "@/lib/constants/landing-data";
import { CheckIcon } from "@/components/icons";
import { SectionHeader } from "@/components/ui/SectionHeader";

export function SimulatorSection() {
  const [incidentStep, setIncidentStep] = useState(0);
  const currentStage = SIMULATOR_STAGES[incidentStep];

  return (
    <section id="simulator" className="py-14 relative z-10 border-t border-neutral-200/80 dark:border-neutral-800/80">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          badge="Smart Threshold Consensus Engine"
          title="Zero 3:00 AM False Alarms. Only Real Incidents."
          description="Transient hiccups happen. PulseWatch cross-verifies failures across multiple consecutive probe intervals before paging your on-call engineers."
        />

        <div className="bg-white dark:bg-[#0a0e1a] p-5 sm:p-6 rounded-2xl border border-neutral-200 dark:border-neutral-800 shadow-sm dark:shadow-2xl transition-colors">
          {/* Step Controls */}
          <div className="flex flex-wrap items-center justify-between gap-3 pb-4 border-b border-neutral-200/80 dark:border-neutral-800/80">
            <div>
              <span className="text-xs text-neutral-500 dark:text-neutral-400">Current Simulator State:</span>
              <span className={`ml-2 text-xs font-mono font-bold px-2 py-0.5 rounded border ${currentStage.badgeColor}`}>
                {currentStage.status}
              </span>
            </div>

            <div className="flex items-center gap-2">
              <button
                onClick={() => setIncidentStep((prev) => (prev > 0 ? prev - 1 : 0))}
                disabled={incidentStep === 0}
                className="px-3 py-1.5 text-xs font-medium rounded-lg bg-neutral-100 hover:bg-neutral-200 dark:bg-neutral-800 dark:hover:bg-neutral-700 text-neutral-700 dark:text-neutral-300 disabled:opacity-40 transition cursor-pointer border border-neutral-200 dark:border-neutral-700"
              >
                ← Prev
              </button>
              <button
                onClick={() => setIncidentStep((prev) => (prev < SIMULATOR_STAGES.length - 1 ? prev + 1 : 0))}
                className="px-3.5 py-1.5 text-xs font-semibold rounded-lg bg-neutral-900 hover:bg-neutral-800 text-white dark:bg-white dark:text-neutral-900 dark:hover:bg-neutral-100 transition cursor-pointer shadow-sm"
              >
                {incidentStep === SIMULATOR_STAGES.length - 1 ? "Reset Simulation" : "Next Failure Step →"}
              </button>
            </div>
          </div>

          {/* Stepper Visual */}
          <div className="grid grid-cols-1 sm:grid-cols-5 gap-2.5 my-5">
            {SIMULATOR_STAGES.map((st, idx) => (
              <button
                key={idx}
                onClick={() => setIncidentStep(idx)}
                className={`p-3 rounded-xl border text-left transition-all text-xs flex flex-col gap-1 cursor-pointer ${
                  incidentStep === idx
                    ? "bg-neutral-100 dark:bg-[#131b2e] border-neutral-400 dark:border-neutral-600 text-neutral-900 dark:text-white shadow-sm ring-1 ring-neutral-300 dark:ring-neutral-700"
                    : "bg-neutral-50/60 dark:bg-[#090d18]/50 border-neutral-200/80 dark:border-neutral-800/80 text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white"
                }`}
              >
                <span className="font-mono text-[10px] text-neutral-400 dark:text-neutral-500">Step {idx + 1}</span>
                <span className="font-semibold text-xs leading-tight">{st.title}</span>
              </button>
            ))}
          </div>

          {/* Stage Detail Box */}
          <div className="bg-neutral-50/80 dark:bg-[#0c101d]/80 rounded-xl p-4 sm:p-5 border border-neutral-200/80 dark:border-neutral-800/80 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="max-w-lg">
              <div className="text-xs font-semibold text-neutral-900 dark:text-slate-200">
                {currentStage.title} — {currentStage.description}
              </div>
              <div className="text-[11px] text-neutral-500 dark:text-slate-400 mt-1">
                Alert Strategy:{" "}
                {currentStage.alertSent ? (
                  <strong className="text-rose-600 dark:text-rose-400">
                    DISPATCHED (RabbitMQ queue: notifications.email + notifications.webhook)
                  </strong>
                ) : (
                  <strong className="text-emerald-600 dark:text-emerald-400">
                    SUPPRESSED (Below consecutive failure threshold of 3)
                  </strong>
                )}
              </div>
            </div>

            <div className="text-xs font-mono px-3 py-2 rounded-lg bg-white dark:bg-black/50 border border-neutral-200 dark:border-neutral-800 text-neutral-700 dark:text-slate-300 shadow-sm">
              {currentStage.alertSent ? (
                <span className="text-rose-600 dark:text-rose-400 flex items-center gap-1.5 font-semibold">
                  <span className="w-2 h-2 rounded-full bg-rose-500 animate-ping" />
                  Mailpit & Webhook Paged
                </span>
              ) : (
                <span className="text-emerald-600 dark:text-emerald-400 flex items-center gap-1.5 font-semibold">
                  <CheckIcon className="w-3.5 h-3.5" />
                  On-call Engineers Sleeping
                </span>
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
