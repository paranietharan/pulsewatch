"use client";

import React, { useState } from "react";
import Link from "next/link";
import * as authApi from "@/lib/api/auth";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";
import { PulseLogoIcon } from "@/components/icons";
import { ThemeToggle } from "@/components/theme/ThemeToggle";
import { CheckCircle2 } from "lucide-react";

export default function SignupPage() {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!firstName.trim() || !email.trim() || !password) {
      setError("Please fill in all required fields.");
      return;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters long.");
      return;
    }

    setSubmitting(true);
    try {
      const res = await authApi.signup({
        first_name: firstName,
        last_name: lastName,
        email,
        password,
      });
      setSuccess(
        res.message ||
          "Account created successfully. Please check your email to verify your account."
      );
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to create account. Please try again.");
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
            Create your account
          </h1>
          <p className="text-xs text-neutral-500 mt-1">
            Start monitoring your endpoints with zero false alarms
          </p>
        </div>

        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800 rounded-xl p-6 shadow-xs">
          {success ? (
            <div className="text-center py-3">
              <div className="w-10 h-10 rounded-full bg-emerald-50 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400 flex items-center justify-center mx-auto mb-3">
                <CheckCircle2 className="w-6 h-6" />
              </div>
              <h2 className="text-sm font-bold text-neutral-900 dark:text-white">
                Verification Email Sent
              </h2>
              <p className="text-xs text-neutral-500 mt-1 leading-relaxed">
                {success}
              </p>
              <div className="mt-5 flex flex-col gap-2">
                <Link
                  href="/login"
                  className="inline-flex items-center justify-center px-4 py-2 text-xs font-semibold rounded-lg bg-neutral-900 text-white dark:bg-white dark:text-neutral-900 hover:opacity-90 transition"
                >
                  Proceed to Sign In
                </Link>
                <Link
                  href="/resend-verification"
                  className="text-[11px] text-neutral-500 hover:text-neutral-900 dark:hover:text-white mt-1"
                >
                  Resend verification email
                </Link>
              </div>
            </div>
          ) : (
            <>
              {error && (
                <div className="mb-4 p-2.5 rounded-lg bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/40 text-rose-700 dark:text-rose-400 text-xs">
                  {error}
                </div>
              )}

              <form onSubmit={handleSubmit} className="flex flex-col gap-3.5">
                <div className="grid grid-cols-2 gap-2.5">
                  <Input
                    label="First name"
                    value={firstName}
                    onChange={(e) => setFirstName(e.target.value)}
                    placeholder="Jane"
                    required
                  />
                  <Input
                    label="Last name"
                    value={lastName}
                    onChange={(e) => setLastName(e.target.value)}
                    placeholder="Doe"
                  />
                </div>

                <Input
                  label="Email address"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="name@example.com"
                  required
                />

                <Input
                  label="Password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Minimum 8 characters"
                  helperText="Must be between 8 and 72 characters"
                  required
                />

                <Button
                  type="submit"
                  variant="primary"
                  size="md"
                  isLoading={submitting}
                  className="w-full mt-2"
                >
                  Create Account
                </Button>
              </form>

              <div className="mt-5 pt-4 border-t border-neutral-100 dark:border-neutral-800/80 text-center text-xs text-neutral-500">
                <span>Already have an account? </span>
                <Link
                  href="/login"
                  className="font-semibold text-neutral-900 dark:text-white hover:underline"
                >
                  Sign in
                </Link>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
