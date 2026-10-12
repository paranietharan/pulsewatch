import React from "react";
import { ARCHITECTURE_PIPELINE } from "@/lib/constants/landing-data";
import { SectionHeader } from "@/components/ui/SectionHeader";

export function ArchitectureSection() {
  return (
    <section id="architecture" className="py-14 relative z-10 border-t border-neutral-200/80 dark:border-neutral-800/80">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          badge="High-Throughput Go Microservices"
          title="Decoupled, Event-Driven Architecture"
          description="Stateless probe workers communicate purely via RabbitMQ message queues. Zero direct database dependencies on notification dispatch."
        />

        {/* Architecture Pipeline */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-3.5 relative">
          {ARCHITECTURE_PIPELINE.map((node, idx) => (
            <div
              key={idx}
              className={`p-4 sm:p-5 rounded-xl border transition-all flex flex-col justify-between ${
                node.isHighlight
                  ? "bg-neutral-50/80 dark:bg-[#0d1424] border-neutral-300 dark:border-neutral-700 shadow-sm"
                  : "bg-white dark:bg-[#0a0e1a] border-neutral-200/80 dark:border-neutral-800/80 shadow-sm hover:border-neutral-300 dark:hover:border-neutral-700"
              }`}
            >
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className={`text-[10px] font-mono px-1.5 py-0.5 rounded border ${node.badgeColor}`}>
                    {node.badge}
                  </span>
                  <span className="text-[11px] font-mono text-neutral-500 dark:text-neutral-400">{node.service}</span>
                </div>
                <h3 className="text-sm font-bold text-neutral-900 dark:text-white font-[family-name:var(--font-heading)]">
                  {node.title}
                </h3>
                <p className="text-xs text-neutral-600 dark:text-neutral-400 mt-1 leading-relaxed">
                  {node.description}
                </p>
              </div>

              <div className="mt-4 pt-3 border-t border-neutral-100 dark:border-neutral-800/80 text-[11px] font-mono text-emerald-600 dark:text-emerald-400">
                {node.connection}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
