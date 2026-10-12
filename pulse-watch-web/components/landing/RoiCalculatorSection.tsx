"use client";

import React, { useState } from "react";
import { SectionHeader } from "@/components/ui/SectionHeader";

const PRESETS = [10000, 25000, 50000, 100000, 250000];

export function RoiCalculatorSection() {
  const [revenueSlider, setRevenueSlider] = useState(50000);

  const costPerMinute = (revenueSlider / (30 * 24 * 60)) * 1.5;
  const annualSaved = costPerMinute * 60 * 3.8;

  return (
    <section id="calculator" className="py-14 relative z-10 border-t border-neutral-200/80 dark:border-neutral-800/80">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          badge="Financial Impact Assessment"
          title="Calculate Your True Cost of Downtime"
          description="Discover how sub-minute outage detection saves thousands in lost revenue and prevents customer churn."
        />

        <div className="bg-white dark:bg-[#0a0e1a] p-6 rounded-2xl border border-neutral-200 dark:border-neutral-800 shadow-sm dark:shadow-2xl transition-colors">
          <div className="grid grid-cols-1 md:grid-cols-12 gap-6 items-center">
            {/* Left: Interactive Slider */}
            <div className="md:col-span-7 flex flex-col gap-4">
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-neutral-600 dark:text-neutral-400">
                  Monthly Recurring Revenue (MRR)
                </span>
                <span className="text-lg font-bold font-mono text-neutral-900 dark:text-white">
                  ${revenueSlider.toLocaleString()} / mo
                </span>
              </div>

              <input
                type="range"
                min="5000"
                max="500000"
                step="5000"
                value={revenueSlider}
                onChange={(e) => setRevenueSlider(Number(e.target.value))}
                className="w-full accent-emerald-600 dark:accent-emerald-400 cursor-pointer h-2 bg-neutral-200 dark:bg-neutral-800 rounded-lg"
              />

              <div className="flex items-center justify-between text-[11px] text-neutral-400 dark:text-neutral-500 font-mono">
                <span>$5,000</span>
                <span>$250,000</span>
                <span>$500,000+</span>
              </div>

              <div className="flex items-center gap-2 pt-2">
                <span className="text-[11px] text-neutral-500 dark:text-neutral-400">Presets:</span>
                {PRESETS.map((val) => {
                  const isActive = revenueSlider === val;
                  return (
                    <button
                      key={val}
                      onClick={() => setRevenueSlider(val)}
                      className={`text-[11px] px-2.5 py-1 rounded font-mono border transition cursor-pointer ${
                        isActive
                          ? "bg-neutral-900 text-white dark:bg-white dark:text-neutral-900 border-neutral-900 dark:border-white shadow-sm"
                          : "bg-neutral-100 dark:bg-neutral-800 text-neutral-600 dark:text-neutral-400 border-neutral-200 dark:border-neutral-700 hover:text-neutral-900 dark:hover:text-white"
                      }`}
                    >
                      ${val / 1000}k
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Right: Calculated Metrics */}
            <div className="md:col-span-5 bg-neutral-50 dark:bg-[#0c101d] p-5 rounded-xl border border-neutral-200/80 dark:border-neutral-800/80 flex flex-col gap-3.5">
              <div>
                <div className="text-[10px] uppercase font-mono text-neutral-500 dark:text-neutral-400">
                  Downtime Cost Per Minute
                </div>
                <div className="text-2xl font-bold font-mono text-rose-600 dark:text-rose-400 mt-0.5">
                  ${costPerMinute.toFixed(2)}
                  <span className="text-xs font-normal text-neutral-500 dark:text-neutral-400"> / min</span>
                </div>
              </div>

              <div className="pt-2 border-t border-neutral-200/80 dark:border-neutral-800/80">
                <div className="text-[10px] uppercase font-mono text-neutral-500 dark:text-neutral-400">
                  Annual Revenue Protected
                </div>
                <div className="text-2xl font-bold font-mono text-emerald-600 dark:text-emerald-400 mt-0.5">
                  ${Math.round(annualSaved).toLocaleString()}
                  <span className="text-xs font-normal text-neutral-500 dark:text-neutral-400"> / yr</span>
                </div>
              </div>

              <div className="text-[10px] text-neutral-500 dark:text-neutral-400 mt-1 leading-normal">
                *Based on catching outages 42 minutes faster than user reports.
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
