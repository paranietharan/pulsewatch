"use client";

import React, { useState } from "react";
import Link from "next/link";
import * as authApi from "@/lib/api/auth";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";
import { PulseLogoIcon } from "@/components/icons";
import { ThemeToggle } from "@/components/theme/ThemeToggle";
import { MailCheck } from "lucide-react";

export default function ResendVerificationPage() {
  const [email, setEmail] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) return;

    setError(null);
    setMessage(null);
    setSubmitting(true);

    try {
      const res = await authApi.resendVerification(email.trim());
      setMessage(
        res.message ||
          "If an unverified account exists with that email address, a new verification link has been sent."
      );
    } catch (err: unknown) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to request verification email."
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-[#07090e] flex flex-col justify-center items-center px-4 py-12 relative">
      <div className="absolute top-4 right-4">
        <ThemeToggle />
      </div>

      <div className="w-full max-w-sm">
        <div className="text-center mb-6">
          <Link href="/" className="inline-flex items-center gap-2 mb-3">
            <div className="w-7 h-7 rounded-lg bg-neutral-100 dark:bg-[#0e1424] border border-neutral-200 dark:border-neutral-800 flex items-center justify-center p-0.5">
              <PulseLogoIcon className="w-full h-full" />
            </div>
            <span className="text-sm font-bold tracking-tight text-neutral-900 dark:text-white">
              PulseWatch
            </span>
          </Link>
          <h1 className="text-lg font-bold tracking-tight text-neutral-900 dark:text-white">
            Resend Verification Email
          </h1>
          <p className="text-xs text-neutral-500 mt-1">
            Enter your email to receive a fresh verification link
          </p>
        </div>

        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800 rounded-xl p-6 shadow-xs">
          {message ? (
            <div className="text-center py-3">
              <div className="w-10 h-10 rounded-full bg-emerald-50 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400 flex items-center justify-center mx-auto mb-3">
                <MailCheck className="w-6 h-6" />
              </div>
              <h2 className="text-sm font-bold text-neutral-900 dark:text-white">
                Request Dispatched
              </h2>
              <p className="text-xs text-neutral-500 mt-1 leading-relaxed">
                {message}
              </p>
              <div className="mt-5">
                <Link
                  href="/login"
                  className="inline-flex items-center justify-center px-4 py-2 text-xs font-semibold rounded-lg bg-neutral-900 text-white dark:bg-white dark:text-neutral-900 hover:opacity-90 transition w-full"
                >
                  Back to Sign In
                </Link>
              </div>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="flex flex-col gap-3.5">
              {error && (
                <div className="p-2.5 rounded-lg bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/40 text-rose-700 dark:text-rose-400 text-xs">
                  {error}
                </div>
              )}

              <Input
                label="Registered email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="name@example.com"
                required
                autoFocus
              />

              <Button
                type="submit"
                variant="primary"
                size="md"
                isLoading={submitting}
                className="w-full mt-1"
              >
                Send Verification Email
              </Button>

              <div className="mt-4 pt-3 border-t border-neutral-100 dark:border-neutral-800/80 text-center text-xs text-neutral-500">
                <Link
                  href="/login"
                  className="font-medium text-neutral-700 dark:text-neutral-300 hover:underline"
                >
                  Return to Sign In
                </Link>
              </div>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
