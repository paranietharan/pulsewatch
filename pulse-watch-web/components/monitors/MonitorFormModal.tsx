"use client";

import React, { useState, useEffect } from "react";
import { Modal } from "@/components/ui/Modal";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { Button } from "@/components/ui/Button";
import { Monitor, HttpMethod, AlertChannel } from "@/types/api";
import * as channelsApi from "@/lib/api/channels";

interface MonitorFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: {
    name: string;
    url: string;
    method: HttpMethod;
    interval_seconds: number;
    timeout_seconds: number;
    expected_status_code: number;
    ssl_check_enabled: boolean;
    failure_threshold: number;
    recovery_threshold: number;
    alert_channel_ids: string[];
  }) => Promise<void>;
  initialData?: Monitor | null;
  title?: string;
}

export function MonitorFormModal({
  isOpen,
  onClose,
  onSubmit,
  initialData,
  title,
}: MonitorFormModalProps) {
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [method, setMethod] = useState<HttpMethod>("GET");
  const [intervalSeconds, setIntervalSeconds] = useState(60);
  const [timeoutSeconds, setTimeoutSeconds] = useState(10);
  const [expectedStatusCode, setExpectedStatusCode] = useState(200);
  const [sslCheckEnabled, setSslCheckEnabled] = useState(true);
  const [failureThreshold, setFailureThreshold] = useState(3);
  const [recoveryThreshold, setRecoveryThreshold] = useState(2);
  const [selectedChannels, setSelectedChannels] = useState<string[]>([]);
  const [channels, setChannels] = useState<AlertChannel[]>([]);

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isOpen) return;

    if (initialData) {
      setName(initialData.name);
      setUrl(initialData.url);
      setMethod(initialData.method);
      setIntervalSeconds(initialData.interval_seconds);
      setTimeoutSeconds(initialData.timeout_seconds);
      setExpectedStatusCode(initialData.expected_status_code);
      setSslCheckEnabled(initialData.ssl_check_enabled);
      setFailureThreshold(initialData.failure_threshold);
      setRecoveryThreshold(initialData.recovery_threshold);
      setSelectedChannels([]);
    } else {
      setName("");
      setUrl("");
      setMethod("GET");
      setIntervalSeconds(60);
      setTimeoutSeconds(10);
      setExpectedStatusCode(200);
      setSslCheckEnabled(true);
      setFailureThreshold(3);
      setRecoveryThreshold(2);
      setSelectedChannels([]);
    }
    setError(null);

    channelsApi.listAlertChannels().then(setChannels).catch(() => setChannels([]));
  }, [isOpen, initialData]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError("Monitor name is required.");
      return;
    }

    if (!url.trim()) {
      setError("Target URL is required.");
      return;
    }

    try {
      const parsed = new URL(url.trim());
      if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
        setError("Target URL must start with http:// or https://");
        return;
      }
    } catch {
      setError("Please enter a valid URL (e.g., https://api.example.com/health)");
      return;
    }

    if (timeoutSeconds > intervalSeconds) {
      setError("Timeout cannot be greater than the check interval.");
      return;
    }

    setSubmitting(true);
    try {
      await onSubmit({
        name: name.trim(),
        url: url.trim(),
        method,
        interval_seconds: Number(intervalSeconds),
        timeout_seconds: Number(timeoutSeconds),
        expected_status_code: Number(expectedStatusCode),
        ssl_check_enabled: sslCheckEnabled,
        failure_threshold: Number(failureThreshold),
        recovery_threshold: Number(recoveryThreshold),
        alert_channel_ids: selectedChannels,
      });
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to save monitor.");
    } finally {
      setSubmitting(false);
    }
  };

  const toggleChannel = (chId: string) => {
    setSelectedChannels((prev) =>
      prev.includes(chId) ? prev.filter((id) => id !== chId) : [...prev, chId]
    );
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={title || (initialData ? "Edit Monitor" : "Create New Monitor")}
      description="Configure target endpoint, probing frequency, and alert thresholds"
      maxWidth="max-w-lg"
    >
      <form onSubmit={handleSubmit} className="flex flex-col gap-3.5">
        {error && (
          <div className="p-2.5 rounded-lg bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/40 text-rose-700 dark:text-rose-400 text-xs">
            {error}
          </div>
        )}

        <Input
          label="Monitor Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Production API Gateway"
          required
          autoFocus
        />

        <div className="grid grid-cols-4 gap-2">
          <div className="col-span-1">
            <Select
              label="Method"
              value={method}
              onChange={(e) => setMethod(e.target.value as HttpMethod)}
              options={[
                { label: "GET", value: "GET" },
                { label: "POST", value: "POST" },
                { label: "HEAD", value: "HEAD" },
                { label: "PUT", value: "PUT" },
                { label: "PATCH", value: "PATCH" },
                { label: "DELETE", value: "DELETE" },
              ]}
            />
          </div>
          <div className="col-span-3">
            <Input
              label="Target URL"
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://api.example.com/health"
              required
            />
          </div>
        </div>

        <div className="grid grid-cols-3 gap-2.5">
          <Input
            label="Interval (sec)"
            type="number"
            min={10}
            max={86400}
            value={intervalSeconds}
            onChange={(e) => setIntervalSeconds(Number(e.target.value))}
            helperText="10s – 86400s"
            required
          />
          <Input
            label="Timeout (sec)"
            type="number"
            min={1}
            max={60}
            value={timeoutSeconds}
            onChange={(e) => setTimeoutSeconds(Number(e.target.value))}
            helperText="1s – 60s"
            required
          />
          <Input
            label="Expected Status"
            type="number"
            min={100}
            max={599}
            value={expectedStatusCode}
            onChange={(e) => setExpectedStatusCode(Number(e.target.value))}
            helperText="e.g. 200"
            required
          />
        </div>

        <div className="grid grid-cols-2 gap-2.5">
          <Input
            label="Failure Threshold"
            type="number"
            min={1}
            max={10}
            value={failureThreshold}
            onChange={(e) => setFailureThreshold(Number(e.target.value))}
            helperText="Consecutive fails before incident"
            required
          />
          <Input
            label="Recovery Threshold"
            type="number"
            min={1}
            max={10}
            value={recoveryThreshold}
            onChange={(e) => setRecoveryThreshold(Number(e.target.value))}
            helperText="Consecutive OKs before resolved"
            required
          />
        </div>

        <div className="flex items-center gap-2 py-1">
          <input
            id="ssl-check"
            type="checkbox"
            checked={sslCheckEnabled}
            onChange={(e) => setSslCheckEnabled(e.target.checked)}
            className="w-4 h-4 rounded text-emerald-600 focus:ring-emerald-500 border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-800 cursor-pointer"
          />
          <label htmlFor="ssl-check" className="text-xs text-neutral-700 dark:text-neutral-300 select-none cursor-pointer">
            Validate SSL/TLS Certificate validity
          </label>
        </div>

        {channels.length > 0 && !initialData && (
          <div className="pt-2 border-t border-neutral-100 dark:border-neutral-800">
            <label className="text-xs font-medium text-neutral-700 dark:text-neutral-300 block mb-1.5">
              Attach Notification Channels
            </label>
            <div className="flex flex-wrap gap-1.5">
              {channels.map((ch) => {
                const isSelected = selectedChannels.includes(ch.id);
                return (
                  <button
                    type="button"
                    key={ch.id}
                    onClick={() => toggleChannel(ch.id)}
                    className={`px-2.5 py-1 text-xs rounded-md border cursor-pointer transition-colors ${
                      isSelected
                        ? "bg-emerald-50 text-emerald-700 border-emerald-300 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-700 font-semibold"
                        : "bg-white text-neutral-600 border-neutral-200 dark:bg-neutral-900 dark:text-neutral-400 dark:border-neutral-800"
                    }`}
                  >
                    {ch.name} ({ch.type})
                  </button>
                );
              })}
            </div>
          </div>
        )}

        <div className="flex items-center justify-end gap-2 pt-3 border-t border-neutral-100 dark:border-neutral-800">
          <Button type="button" variant="secondary" size="md" onClick={onClose} disabled={submitting}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" size="md" isLoading={submitting}>
            {initialData ? "Save Changes" : "Create Monitor"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
