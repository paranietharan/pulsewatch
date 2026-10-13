import React from "react";
import { FEATURES } from "@/lib/constants/landing-data";
import {
  ServerIcon,
  ShieldCheckIcon,
  BoltIcon,
  BellAlertIcon,
  GlobeIcon,
  CodeBracketIcon,
} from "@/components/icons";
import { SectionHeader } from "@/components/ui/SectionHeader";

function renderFeatureIcon(icon: string) {
  switch (icon) {
    case "server":
      return <ServerIcon className="w-4 h-4 text-neutral-800 dark:text-neutral-200" />;
    case "shield":
      return <ShieldCheckIcon className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />;
    case "bolt":
      return <BoltIcon className="w-4 h-4 text-neutral-800 dark:text-neutral-200" />;
    case "bell":
      return <BellAlertIcon className="w-4 h-4 text-amber-600 dark:text-amber-400" />;
    case "globe":
      return <GlobeIcon className="w-4 h-4 text-neutral-800 dark:text-neutral-200" />;
    case "code":
      return <CodeBracketIcon className="w-4 h-4 text-neutral-800 dark:text-neutral-200" />;
    default:
      return <ServerIcon className="w-4 h-4 text-neutral-800 dark:text-neutral-200" />;
  }
}

export function FeaturesSection() {
  return (
    <section id="features" className="py-14 relative z-10 border-t border-neutral-200/80 dark:border-neutral-800/80">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          badge="Engineered For Reliability"
          title="Everything Needed to Keep Services Up"
          description="High-concurrency probing with strict security controls and instant notifications."
        />

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {FEATURES.map((feature, idx) => (
            <div
              key={idx}
              className="bg-white dark:bg-[#0a0e1a] p-5 rounded-2xl border border-neutral-200/80 dark:border-neutral-800/80 hover:border-neutral-300 dark:hover:border-neutral-700 transition-all shadow-sm"
            >
              <div className="w-9 h-9 rounded-lg bg-neutral-100 dark:bg-neutral-800/80 border border-neutral-200 dark:border-neutral-700 flex items-center justify-center mb-3.5">
                {renderFeatureIcon(feature.icon)}
              </div>
              <h3 className="text-sm font-bold text-neutral-900 dark:text-white font-[family-name:var(--font-heading)]">
                {feature.title}
              </h3>
              <p className="text-xs text-neutral-600 dark:text-neutral-400 mt-1.5 leading-relaxed">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
