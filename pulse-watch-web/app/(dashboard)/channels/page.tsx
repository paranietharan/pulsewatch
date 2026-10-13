"use client";

import React, { useState, useEffect, useCallback } from "react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Modal } from "@/components/ui/Modal";
import { EmptyState } from "@/components/ui/EmptyState";
import { DeleteConfirmModal } from "@/components/monitors/DeleteConfirmModal";
import { AlertChannel, AlertChannelType } from "@/types/api";
import * as channelsApi from "@/lib/api/channels";
import {
  Bell,
  Plus,
  Trash2,
  Mail,
  Webhook,
  RefreshCw,
  KeyRound,
  ShieldAlert,
  AlertTriangle,
} from "lucide-react";

export default function AlertChannelsPage() {
  const [channels, setChannels] = useState<AlertChannel[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Modals state
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [deletingChannel, setDeletingChannel] = useState<AlertChannel | null>(null);

  // Form state
  const [channelType, setChannelType] = useState<AlertChannelType>("EMAIL");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [webhookUrl, setWebhookUrl] = useState("");
  const [webhookSecret, setWebhookSecret] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const fetchChannels = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await channelsApi.listAlertChannels();
      setChannels(data || []);
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : "Failed to load alert channels."
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchChannels();
  }, [fetchChannels]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    if (!name.trim()) {
      setFormError("Channel name is required.");
      return;
    }

    let config: Record<string, unknown> = {};

    if (channelType === "EMAIL") {
      if (!email.trim() || !email.includes("@")) {
        setFormError("A valid email address is required.");
        return;
      }
      config = { email: email.trim() };
    } else {
      if (!webhookUrl.trim()) {
        setFormError("Webhook URL is required.");
        return;
      }
      try {
        const u = new URL(webhookUrl.trim());
        if (u.protocol !== "http:" && u.protocol !== "https:") {
          setFormError("Webhook URL must start with http:// or https://");
          return;
        }
      } catch {
        setFormError("Please enter a valid webhook URL.");
        return;
      }
      config = {
        url: webhookUrl.trim(),
        ...(webhookSecret.trim() ? { secret: webhookSecret.trim() } : {}),
      };
    }

    setSubmitting(true);
    try {
      await channelsApi.createAlertChannel({
        type: channelType,
        name: name.trim(),
        config,
      });
      setCreateModalOpen(false);
      setName("");
      setEmail("");
      setWebhookUrl("");
      setWebhookSecret("");
      await fetchChannels();
    } catch (err: unknown) {
      setFormError(
        err instanceof Error ? err.message : "Failed to create alert channel."
      );
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deletingChannel) return;
    await channelsApi.deleteAlertChannel(deletingChannel.id);
    setDeletingChannel(null);
    await fetchChannels();
  };

  return (
    <DashboardLayout
      title="Alert Channels"
      description="Configure email dispatches and HTTP webhook endpoints for incident notifications"
      action={
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchChannels}
            title="Refresh channels"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="primary"
            size="sm"
            onClick={() => setCreateModalOpen(true)}
          >
            <Plus className="w-3.5 h-3.5" />
            <span>Add Channel</span>
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-4">
        {/* Error Banner */}
        {error && (
          <div className="p-4 rounded-xl bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div className="flex items-center gap-2.5">
              <AlertTriangle className="w-4 h-4 text-rose-600 dark:text-rose-400 shrink-0" />
              <div>
                <p className="text-xs font-semibold text-rose-800 dark:text-rose-300">
                  Failed to load notification channels
                </p>
                <p className="text-[11px] text-rose-600 dark:text-rose-400 mt-0.5">
                  {error}
                </p>
              </div>
            </div>
            <Button variant="secondary" size="sm" onClick={fetchChannels}>
              <RefreshCw className="w-3.5 h-3.5 mr-1" />
              Retry
            </Button>
          </div>
        )}

        {/* Notice on Notification Service */}
        <div className="p-3.5 rounded-xl bg-neutral-100 dark:bg-neutral-900/60 border border-neutral-200 dark:border-neutral-800/80 flex items-start gap-3">
          <ShieldAlert className="w-4 h-4 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" />
          <div className="text-xs text-neutral-600 dark:text-neutral-400 leading-relaxed">
            <span className="font-semibold text-neutral-900 dark:text-white">
              Automated Incident Alerts:
            </span>{" "}
            When an endpoint breaches its failure threshold, the background notification service dispatches real-time alerts through your attached channels.
          </div>
        </div>

        {/* Channels Grid / Table */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl shadow-2xs overflow-hidden">
          {error ? (
            <div className="py-12 text-center text-xs text-neutral-500">
              Could not display notification channels due to a server error.
            </div>
          ) : channels.length === 0 && !loading ? (
            <div className="py-12">
              <EmptyState
                icon={Bell}
                title="No alert channels configured"
                description="Create an email or webhook channel so your team receives instant notifications during service downtime."
                action={
                  <Button
                    variant="primary"
                    size="sm"
                    onClick={() => setCreateModalOpen(true)}
                  >
                    <Plus className="w-3.5 h-3.5 mr-1" />
                    Add First Channel
                  </Button>
                }
              />
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse text-xs">
                <thead>
                  <tr className="border-b border-neutral-200 dark:border-neutral-800 text-[11px] font-semibold text-neutral-500 uppercase tracking-wider bg-neutral-50/50 dark:bg-neutral-900/30">
                    <th className="py-3 px-3.5">Type</th>
                    <th className="py-3 px-3.5">Channel Name</th>
                    <th className="py-3 px-3.5">Destination / Config</th>
                    <th className="py-3 px-3.5">Created</th>
                    <th className="py-3 px-3.5 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-100 dark:divide-neutral-800/60">
                  {channels.map((ch) => {
                    const isEmail = ch.type === "EMAIL";
                    const destEmail = isEmail
                      ? (ch.config?.email as string) || "—"
                      : null;
                    const destUrl = !isEmail
                      ? (ch.config?.url as string) || "—"
                      : null;
                    const hasSecret = !isEmail && Boolean(ch.config?.secret);

                    return (
                      <tr
                        key={ch.id}
                        className="hover:bg-neutral-50/70 dark:hover:bg-neutral-900/40 transition-colors"
                      >
                        <td className="py-3 px-3.5 whitespace-nowrap">
                          <span
                            className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[11px] font-semibold border ${
                              isEmail
                                ? "bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900/60"
                                : "bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-950/40 dark:text-purple-300 dark:border-purple-900/60"
                            }`}
                          >
                            {isEmail ? (
                              <Mail className="w-3 h-3" />
                            ) : (
                              <Webhook className="w-3 h-3" />
                            )}
                            {ch.type}
                          </span>
                        </td>
                        <td className="py-3 px-3.5 font-semibold text-neutral-900 dark:text-white whitespace-nowrap">
                          {ch.name}
                        </td>
                        <td className="py-3 px-3.5 font-mono text-[11px] text-neutral-600 dark:text-neutral-400 max-w-sm truncate">
                          {isEmail ? (
                            <span>{destEmail}</span>
                          ) : (
                            <div className="flex items-center gap-2">
                              <span className="truncate">{destUrl}</span>
                              {hasSecret && (
                                <span
                                  className="inline-flex items-center gap-1 text-[10px] text-neutral-500 bg-neutral-100 dark:bg-neutral-800 px-1.5 py-0.5 rounded"
                                  title="Signing secret is securely masked"
                                >
                                  <KeyRound className="w-2.5 h-2.5" />
                                  ••••••••
                                </span>
                              )}
                            </div>
                          )}
                        </td>
                        <td className="py-3 px-3.5 font-mono text-[11px] text-neutral-500 whitespace-nowrap">
                          {new Date(ch.created_at).toLocaleDateString()}
                        </td>
                        <td className="py-3 px-3.5 text-right whitespace-nowrap">
                          <button
                            onClick={() => setDeletingChannel(ch)}
                            title="Delete alert channel"
                            className="p-1.5 rounded-md text-xs text-neutral-400 hover:bg-rose-50 dark:hover:bg-rose-950/40 hover:text-rose-600 dark:hover:text-rose-400 transition-colors"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Create Channel Modal */}
      <Modal
        isOpen={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        title="Add Notification Channel"
        description="Choose channel type and enter target destination details"
        maxWidth="max-w-md"
      >
        <form onSubmit={handleCreate} className="flex flex-col gap-3.5">
          {formError && (
            <div className="p-2.5 rounded-lg bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/40 text-rose-700 dark:text-rose-400 text-xs">
              {formError}
            </div>
          )}

          {/* Type selector buttons */}
          <div>
            <label className="text-xs font-medium text-neutral-700 dark:text-neutral-300 block mb-1.5">
              Channel Type
            </label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setChannelType("EMAIL")}
                className={`py-2 px-3 rounded-lg border text-xs font-semibold flex items-center justify-center gap-2 transition-colors cursor-pointer ${
                  channelType === "EMAIL"
                    ? "bg-emerald-50 text-emerald-800 border-emerald-300 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-700"
                    : "bg-white text-neutral-600 border-neutral-200 dark:bg-neutral-900 dark:text-neutral-400 dark:border-neutral-800"
                }`}
              >
                <Mail className="w-3.5 h-3.5" />
                <span>Email Address</span>
              </button>

              <button
                type="button"
                onClick={() => setChannelType("WEBHOOK")}
                className={`py-2 px-3 rounded-lg border text-xs font-semibold flex items-center justify-center gap-2 transition-colors cursor-pointer ${
                  channelType === "WEBHOOK"
                    ? "bg-emerald-50 text-emerald-800 border-emerald-300 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-700"
                    : "bg-white text-neutral-600 border-neutral-200 dark:bg-neutral-900 dark:text-neutral-400 dark:border-neutral-800"
                }`}
              >
                <Webhook className="w-3.5 h-3.5" />
                <span>HTTP Webhook</span>
              </button>
            </div>
          </div>

          <Input
            label="Channel Name"
            placeholder={
              channelType === "EMAIL" ? "On-Call Alerts Email" : "Slack Inbound Webhook"
            }
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            autoFocus
          />

          {channelType === "EMAIL" ? (
            <Input
              label="Recipient Email Address"
              type="email"
              placeholder="alerts@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              helperText="Notifications will be dispatched directly to this address."
              required
            />
          ) : (
            <>
              <Input
                label="Webhook Endpoint URL"
                type="url"
                placeholder="https://hooks.example.com/services/alerts"
                value={webhookUrl}
                onChange={(e) => setWebhookUrl(e.target.value)}
                helperText="Must be a valid HTTP or HTTPS endpoint."
                required
              />

              <Input
                label="HMAC Signing Secret (Optional)"
                type="password"
                placeholder="whsec_..."
                value={webhookSecret}
                onChange={(e) => setWebhookSecret(e.target.value)}
                helperText="Used to compute sha256 signature for payload verification."
              />
            </>
          )}

          <div className="flex items-center justify-end gap-2 pt-3 border-t border-neutral-100 dark:border-neutral-800">
            <Button
              type="button"
              variant="secondary"
              size="md"
              onClick={() => setCreateModalOpen(false)}
              disabled={submitting}
            >
              Cancel
            </Button>
            <Button type="submit" variant="primary" size="md" isLoading={submitting}>
              Save Channel
            </Button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation Modal */}
      <DeleteConfirmModal
        isOpen={Boolean(deletingChannel)}
        onClose={() => setDeletingChannel(null)}
        onConfirm={handleDelete}
        title="Delete Alert Channel"
        message="Are you sure you want to remove this notification channel? Monitors linked to this channel will no longer dispatch alerts here."
        itemName={deletingChannel?.name}
      />
    </DashboardLayout>
  );
}
