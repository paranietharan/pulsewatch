import React from "react";

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export function Input({
  label,
  error,
  helperText,
  id,
  className = "",
  ...props
}: InputProps) {
  const inputId = id || (label ? label.toLowerCase().replace(/\s+/g, "-") : undefined);

  return (
    <div className="flex flex-col gap-1.5 w-full">
      {label && (
        <label
          htmlFor={inputId}
          className="text-xs font-medium text-neutral-700 dark:text-neutral-300"
        >
          {label}
        </label>
      )}
      <input
        id={inputId}
        className={`w-full px-3 py-2 text-xs rounded-lg border bg-white dark:bg-[#0c101d] text-neutral-900 dark:text-neutral-100 placeholder-neutral-400 dark:placeholder-neutral-600 focus:outline-none focus:ring-1 transition-colors ${
          error
            ? "border-rose-400 dark:border-rose-600 focus:ring-rose-500"
            : "border-neutral-200 dark:border-neutral-800 focus:border-neutral-400 dark:focus:border-neutral-600 focus:ring-neutral-400"
        } ${className}`}
        {...props}
      />
      {error && <p className="text-[11px] text-rose-600 dark:text-rose-400">{error}</p>}
      {!error && helperText && (
        <p className="text-[11px] text-neutral-500 dark:text-neutral-500">{helperText}</p>
      )}
    </div>
  );
}
