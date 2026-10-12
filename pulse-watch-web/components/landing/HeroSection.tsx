import React from "react";
import { ArrowRightIcon, CodeBracketIcon } from "@/components/icons";
import { HERO_METRICS } from "@/lib/constants/landing-data";

export function HeroSection() {
  return (
    <section className="relative pt-20 pb-12 sm:pt-28 sm:pb-16 overflow-hidden">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 text-center relative z-10">
        {/* Micro Pill */}
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-neutral-100 dark:bg-neutral-900/80 border border-neutral-200 dark:border-neutral-800 text-[11px] font-medium text-neutral-700 dark:text-neutral-300 mb-6 transition-colors">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
          <span>High-Throughput Go Workers · SSRF Shield · Zero False Positives</span>
        </div>

        {/* Minimal Headline */}
        <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold tracking-tight text-neutral-900 dark:text-white max-w-3xl mx-auto leading-[1.12]">
          Distributed uptime probing.
          <span className="block mt-1 text-neutral-400 dark:text-neutral-500 font-semibold">
            Engineered for zero false alarms.
          </span>
        </h1>

        {/* Subtitle */}
        <p className="text-sm sm:text-base text-neutral-600 dark:text-neutral-400 max-w-2xl mx-auto mt-4 leading-relaxed">
          Continuous HTTP, TCP, and TLS probing powered by containerized Go workers.
          Microsecond telemetry, RabbitMQ-driven event streaming, and 3-step consensus
          to eliminate alert fatigue before waking your team.
        </p>

        {/* Minimal Action CTAs */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-3 mt-7">
          <a
            href="#demo"
            className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl font-semibold text-white bg-neutral-900 hover:bg-neutral-800 dark:bg-white dark:text-neutral-900 dark:hover:bg-neutral-100 transition-all text-xs w-full sm:w-auto shadow-sm"
          >
            <span>Launch Live Console</span>
            <ArrowRightIcon className="w-3.5 h-3.5" />
          </a>

          <a
            href="#architecture"
            className="inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl font-medium text-neutral-700 dark:text-neutral-300 bg-neutral-100/80 hover:bg-neutral-200/70 dark:bg-neutral-900/60 dark:hover:bg-neutral-800 border border-neutral-200 dark:border-neutral-800 transition-all text-xs w-full sm:w-auto"
          >
            <span>View Architecture</span>
          </a>

          <a
            href="https://github.com"
            target="_blank"
            rel="noreferrer"
            className="text-xs text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white px-3 py-2 flex items-center gap-1.5 transition-colors"
          >
            <CodeBracketIcon className="w-4 h-4 text-neutral-400" />
            <span>GitHub</span>
          </a>
        </div>

        {/* Minimal 4-Metric Ribbon */}
        <div className="mt-12 max-w-3xl mx-auto grid grid-cols-2 md:grid-cols-4 gap-3 text-left">
          {HERO_METRICS.map((metric, idx) => (
            <div
              key={idx}
              className="bg-neutral-50 dark:bg-[#0b0f19] border border-neutral-200/80 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-sm transition-colors"
            >
              <div className="text-xl font-bold font-mono text-neutral-900 dark:text-white">
                {metric.value}
              </div>
              <div className="text-[11px] text-neutral-500 dark:text-neutral-400 font-medium mt-0.5">
                {metric.label}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
