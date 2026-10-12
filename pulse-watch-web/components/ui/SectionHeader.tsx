import React from "react";

interface SectionHeaderProps {
  badge: string;
  title: string;
  description: string;
  className?: string;
}

export function SectionHeader({
  badge,
  title,
  description,
  className = "max-w-xl mx-auto mb-10 text-center",
}: SectionHeaderProps) {
  return (
    <div className={className}>
      <span className="text-[11px] font-mono font-semibold uppercase tracking-wider text-emerald-600 dark:text-emerald-400">
        {badge}
      </span>
      <h2 className="text-2xl sm:text-3xl font-bold tracking-tight text-neutral-900 dark:text-white mt-1.5 font-[family-name:var(--font-heading)]">
        {title}
      </h2>
      <p className="text-xs sm:text-sm text-neutral-600 dark:text-neutral-400 mt-2 leading-relaxed">
        {description}
      </p>
    </div>
  );
}
