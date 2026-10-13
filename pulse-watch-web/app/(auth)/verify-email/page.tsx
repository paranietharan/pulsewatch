"use client";

import React, { useState, useEffect, Suspense } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import * as authApi from "@/lib/api/auth";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { PulseLogoIcon } from "@/components/icons";
import { ThemeToggle } from "@/components/theme/ThemeToggle";
import { CheckCircle2, AlertCircle, Loader2 } from "lucide-react";

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const tokenParam = searchParams.get("token") || "";

  const [token, setToken] = useState(tokenParam);
  const [status, setStatus] = useState<"idle" | "verifying" | "success" | "error">(
    tokenParam ? "verifying" : "idle"
  );
  const [message, setMessage] = useState<string>("");

  useEffect(() => {
    if (!tokenParam) return;

    let isMounted = true;
    (async () => {
      try {
        const res = await authApi.verifyEmail(tokenParam);
        if (isMounted) {
          setStatus("success");
          setMessage(res.message || "Email address verified successfully.");
        }
      } catch (err: unknown) {
        if (isMounted) {
          setStatus("error");
          setMessage(
            err instanceof Error
              ? err.message
              : "Invalid or expired verification token."
          );
        }
      }
    })();

    return () => {
      isMounted = false;
    };
  }, [tokenParam]);

  const handleManualSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token.trim()) return;

    setStatus("verifying");
    try {
      const res = await authApi.verifyEmail(token.trim());
      setStatus("success");
      setMessage(res.message || "Email address verified successfully.");
    } catch (err: unknown) {
      setStatus("error");
      setMessage(
        err instanceof Error
          ? err.message
          : "Invalid or expired verification token."
      );
    }
  };

  return (
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
          Email Verification
        </h1>
        <p className="text-xs text-neutral-500 mt-1">
          Verify your account to enable probe monitoring
        </p>
      </div>

      <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800 rounded-xl p-6 shadow-xs text-center">
        {status === "verifying" && (
          <div className="py-6 flex flex-col items-center">
            <Loader2 className="w-8 h-8 text-neutral-400 animate-spin mb-3" />
            <p className="text-xs text-neutral-600 dark:text-neutral-400 font-medium">
              Verifying your account token...
            </p>
          </div>
        )}

        {status === "success" && (
          <div className="py-3">
            <div className="w-10 h-10 rounded-full bg-emerald-50 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400 flex items-center justify-center mx-auto mb-3">
              <CheckCircle2 className="w-6 h-6" />
            </div>
            <h2 className="text-sm font-bold text-neutral-900 dark:text-white">
              Verified Successfully
            </h2>
            <p className="text-xs text-neutral-500 mt-1 leading-relaxed">
              {message}
            </p>
            <div className="mt-5">
              <Link
                href="/login"
                className="inline-flex items-center justify-center px-4 py-2 text-xs font-semibold rounded-lg bg-neutral-900 text-white dark:bg-white dark:text-neutral-900 hover:opacity-90 transition w-full"
              >
                Sign In to Dashboard
              </Link>
            </div>
          </div>
        )}

        {status === "error" && (
          <div className="py-3">
            <div className="w-10 h-10 rounded-full bg-rose-50 dark:bg-rose-950/40 text-rose-600 dark:text-rose-400 flex items-center justify-center mx-auto mb-3">
              <AlertCircle className="w-6 h-6" />
            </div>
            <h2 className="text-sm font-bold text-neutral-900 dark:text-white">
              Verification Failed
            </h2>
            <p className="text-xs text-neutral-500 mt-1 leading-relaxed">
              {message}
            </p>
            <div className="mt-5 flex flex-col gap-2">
              <Link
                href="/resend-verification"
                className="inline-flex items-center justify-center px-4 py-2 text-xs font-semibold rounded-lg bg-neutral-900 text-white dark:bg-white dark:text-neutral-900 hover:opacity-90 transition w-full"
              >
                Request New Verification Link
              </Link>
              <Link
                href="/login"
                className="text-xs text-neutral-500 hover:text-neutral-900 dark:hover:text-white mt-1"
              >
                Back to Sign In
              </Link>
            </div>
          </div>
        )}

        {status === "idle" && (
          <form onSubmit={handleManualSubmit} className="flex flex-col gap-3 text-left">
            <Input
              label="Verification token"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="Paste token from verification email"
              required
            />
            <Button type="submit" variant="primary" size="md" className="w-full mt-2">
              Verify Account
            </Button>
            <div className="text-center mt-3">
              <Link
                href="/resend-verification"
                className="text-[11px] text-neutral-500 hover:text-neutral-900 dark:hover:text-white"
              >
                Lost your link? Resend email
              </Link>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-[#07090e] flex flex-col justify-center items-center px-4 py-12 relative">
      <div className="absolute top-4 right-4">
        <ThemeToggle />
      </div>
      <Suspense fallback={<Loader2 className="w-6 h-6 animate-spin text-neutral-400" />}>
        <VerifyEmailContent />
      </Suspense>
    </div>
  );
}
