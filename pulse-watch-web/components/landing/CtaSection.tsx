import React from "react";

export function CtaSection() {
  return (
    <section id="cta" className="py-14 relative z-10 border-t border-neutral-200/80 dark:border-neutral-800/80">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <div className="bg-neutral-50/80 dark:bg-[#0a0e1a] p-8 sm:p-10 rounded-2xl border border-neutral-200 dark:border-neutral-800 shadow-sm transition-colors">
          <h2 className="text-2xl sm:text-3xl font-bold tracking-tight text-neutral-900 dark:text-white font-[family-name:var(--font-heading)]">
            Know when services go down before your customers do.
          </h2>
          <p className="text-xs sm:text-sm text-neutral-600 dark:text-neutral-400 max-w-xl mx-auto mt-2 leading-relaxed">
            Launch PulseWatch in under 60 seconds with Docker Compose. Zero proprietary lock-in.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3 mt-6">
            <a
              href="#pricing"
              className="inline-flex items-center justify-center px-5 py-2.5 rounded-xl font-semibold text-white bg-neutral-900 hover:bg-neutral-800 dark:bg-white dark:text-neutral-900 dark:hover:bg-neutral-100 transition-all text-xs shadow-sm"
            >
              Get Started Now
            </a>
            <a
              href="http://localhost:8080/health"
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center justify-center px-4 py-2.5 rounded-xl font-medium text-neutral-700 dark:text-neutral-300 bg-white hover:bg-neutral-100 dark:bg-neutral-900/60 dark:hover:bg-neutral-800 border border-neutral-200 dark:border-neutral-800 transition-all text-xs"
            >
              Inspect API Health
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
