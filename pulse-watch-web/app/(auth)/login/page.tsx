"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";
import { PulseLogoIcon } from "@/components/icons";
import { ThemeToggle } from "@/components/theme/ThemeToggle";

export default function LoginPage() {
  const { login, isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace("/dashboard");
    }
  }, [isLoading, isAuthenticated, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!email.trim() || !password) {
      setError("Please enter both email and password.");
      return;
    }

    setSubmitting(true);
    try {
      await login(email, password);
      router.push("/dashboard");
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to log in. Please check your credentials.");
      }
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
            Sign in to your account
          </h1>
          <p className="text-xs text-neutral-500 mt-1">
            Enter your credentials to access the monitoring cockpit
          </p>
        </div>

        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800 rounded-xl p-6 shadow-xs">
          {error && (
            <div className="mb-4 p-2.5 rounded-lg bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/40 text-rose-700 dark:text-rose-400 text-xs">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <Input
              label="Email address"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="name@example.com"
              required
              autoFocus
            />

            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label className="text-xs font-medium text-neutral-700 dark:text-neutral-300">
                  Password
                </label>
                <Link
                  href="/forgot-password"
                  className="text-[11px] text-neutral-500 hover:text-neutral-900 dark:hover:text-white"
                >
                  Forgot password?
                </Link>
              </div>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
              />
            </div>

            <Button
              type="submit"
              variant="primary"
              size="md"
              isLoading={submitting}
              className="w-full mt-1"
            >
              Sign In
            </Button>
          </form>

          <div className="mt-5 pt-4 border-t border-neutral-100 dark:border-neutral-800/80 text-center text-xs text-neutral-500">
            <span>Don&apos;t have an account? </span>
            <Link
              href="/signup"
              className="font-semibold text-neutral-900 dark:text-white hover:underline"
            >
              Sign up
            </Link>
          </div>
        </div>

        <div className="text-center mt-4">
          <Link
            href="/resend-verification"
            className="text-[11px] text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-300"
          >
            Didn&apos;t receive verification email?
          </Link>
        </div>
      </div>
    </div>
  );
}
