import React from "react";
import { PRICING_PLANS } from "@/lib/constants/landing-data";
import { CheckIcon } from "@/components/icons";
import { SectionHeader } from "@/components/ui/SectionHeader";

export function PricingSection() {
  return (
    <section id="pricing" className="py-14 relative z-10 border-t border-neutral-200/80 dark:border-neutral-800/80">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
        <SectionHeader
          badge="Transparent Pricing"
          title="Start Free. Scale as You Grow."
          description="No hidden check limits or per-seat billing surprises."
        />

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-stretch">
          {PRICING_PLANS.map((plan) => {
            if (plan.isPopular) {
              return (
                <div
                  key={plan.id}
                  className="bg-white dark:bg-[#0a0e1a] p-5 rounded-2xl border-2 border-neutral-900 dark:border-neutral-300 shadow-md flex flex-col justify-between relative transition-colors"
                >
                  <span className="absolute -top-3 right-4 text-[10px] font-bold font-mono px-2.5 py-0.5 rounded-full bg-neutral-900 text-white dark:bg-white dark:text-neutral-900 uppercase shadow-sm">
                    Most Popular
                  </span>

                  <div>
                    <span className="text-xs font-bold text-neutral-900 dark:text-white">{plan.name}</span>
                    <div className="text-2xl font-bold font-mono text-neutral-900 dark:text-white mt-2">
                      ${plan.price}{" "}
                      <span className="text-xs font-normal text-neutral-500 dark:text-neutral-400">{plan.period}</span>
                    </div>
                    <p className="text-xs text-neutral-600 dark:text-neutral-400 mt-1">{plan.tagline}</p>

                    <ul className="mt-4 flex flex-col gap-2 text-xs text-neutral-700 dark:text-neutral-300">
                      {plan.features.map((feat, i) => (
                        <li key={i} className="flex items-center gap-2">
                          <CheckIcon className="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
                          <span>{feat}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  <a
                    href="#cta"
                    className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl font-semibold text-white bg-neutral-900 hover:bg-neutral-800 dark:bg-white dark:text-neutral-900 dark:hover:bg-neutral-100 transition-all text-xs w-full mt-6 shadow-sm"
                  >
                    {plan.ctaText}
                  </a>
                </div>
              );
            }

            return (
              <div
                key={plan.id}
                className="bg-white dark:bg-[#0a0e1a] p-5 rounded-2xl border border-neutral-200/80 dark:border-neutral-800/80 shadow-sm flex flex-col justify-between transition-colors"
              >
                <div>
                  <span className="text-xs font-bold text-neutral-800 dark:text-slate-300">{plan.name}</span>
                  <div className="text-2xl font-bold font-mono text-neutral-900 dark:text-white mt-2">
                    ${plan.price}{" "}
                    <span className="text-xs font-normal text-neutral-500 dark:text-neutral-400">{plan.period}</span>
                  </div>
                  <p className="text-xs text-neutral-600 dark:text-neutral-400 mt-1">{plan.tagline}</p>

                  <ul className="mt-4 flex flex-col gap-2 text-xs text-neutral-600 dark:text-neutral-400">
                    {plan.features.map((feat, i) => (
                      <li key={i} className="flex items-center gap-2">
                        <CheckIcon className="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
                        <span>{feat}</span>
                      </li>
                    ))}
                  </ul>
                </div>

                <a
                  href="#cta"
                  className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl font-medium text-neutral-700 dark:text-neutral-300 bg-neutral-100/80 hover:bg-neutral-200/70 dark:bg-neutral-900/60 dark:hover:bg-neutral-800 border border-neutral-200 dark:border-neutral-800 transition-all text-xs w-full mt-6"
                >
                  {plan.ctaText}
                </a>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
